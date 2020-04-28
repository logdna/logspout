// Package adapter is the implementation of LogDNA LogSpout Adapter
package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/gliderlabs/logspout/router"
	"github.com/gojektech/heimdall"
	"github.com/gojektech/heimdall/httpclient"
)

// New method of Adapter:
func New(config Configuration) *Adapter {
	backoff := heimdall.NewConstantBackoff(config.BackoffInterval, config.JitterInterval)
	retrier := heimdall.NewRetrier(backoff)
	httpClient := httpclient.NewClient(
		httpclient.WithHTTPTimeout(config.HTTPTimeout),
		httpclient.WithRetrier(retrier),
		httpclient.WithRetryCount(int(config.RequestRetryCount)),
	)
	adapterVersion := os.Getenv("BUILD_VERSION")

	adapter := &Adapter{
		Config:        config,
		HTTPClient:    httpClient,
		Logger:        log.New(os.Stdout, config.Hostname+" ", log.LstdFlags),
		Queue:         make(chan Line),
		instrumenting: newInstrumentingAdapter(),
		version:       adapterVersion,
	}

	go adapter.instrumenting.fireBuildInfo(adapterVersion)

	go adapter.readQueue()
	return adapter
}

// getLevel method is for specifying the Level:
func (adapter *Adapter) getLevel(source string) string {
	switch source {
	case "stdout":
		return "INFO"
	case "stderr":
		return "ERROR"
	}
	return ""
}

// getHost method is for deciding what to choose as a hostname:
func (adapter *Adapter) getHost(containerHostname string) string {
	host := containerHostname
	if adapter.Config.Hostname != "" {
		host = adapter.Config.Hostname
	}
	return host
}

// getTags method is for extracting the tags from templates:
func (adapter *Adapter) getTags(m *router.Message) string {

	if adapter.Config.Tags == "" {
		return ""
	}

	splitTags := strings.Split(adapter.Config.Tags, ",")
	var listTags []string
	existenceMap := map[string]bool{}
	for _, t := range splitTags {
		parsed := false
		if matched, error := regexp.Match(`{{.+}}`, []byte(t)); matched && error == nil {
			var parsedTagBytes bytes.Buffer
			tmp, e := template.New("parsedTag").Parse(t)
			if e == nil {
				err := tmp.ExecuteTemplate(&parsedTagBytes, "parsedTag", m)
				if err == nil {
					parsedTag := parsedTagBytes.String()
					for _, p := range strings.Split(parsedTag, ":") {
						if !existenceMap[p] {
							listTags = append(listTags, p)
							existenceMap[p] = true
						}
					}
					parsed = true
				}
			}
		}

		if !parsed && !existenceMap[t] {
			listTags = append(listTags, t)
			existenceMap[t] = true
		}
	}

	return strings.Join(listTags, ",")
}

// Stream method is for streaming the messages:
func (adapter *Adapter) Stream(logstream chan *router.Message) {
	defer func(begin time.Time) {
		adapter.instrumenting.fireAddRequestCount(streamActionLabel)
		adapter.instrumenting.fireAddRequestLatency(begin, streamActionLabel)
	}(time.Now())

	for m := range logstream {
		if m.Data == "" {
			continue
		}

		messageStr, err := json.Marshal(Message{
			Message: m.Data,
			Container: ContainerInfo{
				Name: strings.Trim(m.Container.Name, "/"),
				ID:   m.Container.ID,
				Config: ContainerConfig{
					Image:    m.Container.Config.Image,
					Hostname: m.Container.Config.Hostname,
					Labels:   m.Container.Config.Labels,
				},
			},
			Level:    adapter.getLevel(m.Source),
			Hostname: adapter.getHost(m.Container.Config.Hostname),
			Tags:     adapter.getTags(m),
		})

		if err != nil {
			go adapter.instrumenting.fireAddStreamMarshalDataError()
			adapter.Logger.Println(
				fmt.Errorf(
					"JSON Marshalling Error: %s. Missing message data: %s",
					err.Error(),
					m.Data,
				),
			)
		} else {
			adapter.Queue <- Line{
				Line:      string(messageStr),
				File:      m.Container.Name,
				Timestamp: time.Now().Unix(),
			}
		}
	}
}

// readQueue is a method for reading from queue:
func (adapter *Adapter) readQueue() {
	defer func(begin time.Time) {
		adapter.instrumenting.fireAddRequestCount(readQueueActionLabel)
		adapter.instrumenting.fireAddRequestLatency(begin, readQueueActionLabel)
	}(time.Now())

	buffer := make([]Line, 0)
	bufferSize := 0

	timeout := time.NewTimer(adapter.Config.FlushInterval)

	for {
		select {
		case msg := <-adapter.Queue:
			if bufferSize >= int(adapter.Config.MaxBufferSize) {
				timeout.Stop()
				adapter.flushBuffer(buffer)
				buffer = make([]Line, 0)
				bufferSize = 0
			}

			buffer = append(buffer, msg)
			bufferSize += len(msg.Line)

		case <-timeout.C:
			if bufferSize > 0 {
				adapter.flushBuffer(buffer)
				buffer = make([]Line, 0)
				bufferSize = 0
			}
		}

		timeout.Reset(adapter.Config.FlushInterval)
	}
}

// flushBuffer is a method for flushing the lines:
func (adapter *Adapter) flushBuffer(buffer []Line) {
	defer func(begin time.Time) {
		adapter.instrumenting.fireAddRequestCount(flushBufferActionLabel)
		adapter.instrumenting.fireAddRequestLatency(begin, flushBufferActionLabel)
	}(time.Now())

	var data bytes.Buffer

	body := struct {
		Lines []Line `json:"lines"`
	}{
		Lines: buffer,
	}

	if error := json.NewEncoder(&data).Encode(body); error != nil {
		go adapter.instrumenting.fireAddFlushBufferEncodeDataError()
		adapter.Logger.Println(
			fmt.Errorf(
				"JSON Encoding Error: %s",
				error.Error(),
			),
		)
		return
	}

	urlValues := url.Values{}
	urlValues.Add("hostname", "logdna_logspout")
	url := "https://" + adapter.Config.LogDNAURL + "?" + urlValues.Encode()
	req, _ := http.NewRequest(http.MethodPost, url, &data)
	req.Header.Set("user-agent", "logspout/"+adapter.version)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.SetBasicAuth(adapter.Config.LogDNAKey, "")
	resp, err := adapter.HTTPClient.Do(req)

	if err != nil {
		go adapter.instrumenting.fireAddFlushBufferLogDNACodeError()
		adapter.Logger.Println(
			fmt.Errorf(
				"HTTP Client Post Request Error: %s",
				err.Error(),
			),
		)
		return
	}

	if resp != nil {
		go adapter.instrumenting.fireAddLogDNAClientRequest(flushBufferActionLabel, fmt.Sprint(resp.StatusCode))

		if resp.StatusCode != http.StatusOK {
			adapter.Logger.Println(
				fmt.Errorf(
					"Received Status Code: %d While Sending Message",
					resp.StatusCode,
				),
			)
		}
		defer resp.Body.Close()
	} else {
		go adapter.instrumenting.fireAddFlushBufferLogDNACodeError()
		adapter.Logger.Println(
			fmt.Errorf(
				"HTTP Client Post Request Error: %s",
				"request response is nil",
			),
		)
	}
}
