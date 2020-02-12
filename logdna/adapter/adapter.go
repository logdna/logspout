// Package adapter is the implementation of LogDNA LogSpout Adapter
package adapter

import (
    "bytes"
    "encoding/binary"
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

    adapter := &Adapter{
        Config:     config,
        Queue:      make(chan Line),
        HTTPClient: httpClient,
    }

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
    if (adapter.Config.Hostname != "") {
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
    for m := range logstream {
        if m.Data == "" {
            continue
        }

        messageStr, err := json.Marshal(Message{
            Message:    m.Data,
            Container:  ContainerInfo{
                Name:   m.Container.Name,
                ID:     m.Container.ID,
                Config: ContainerConfig{
                    Image:      m.Container.Config.Image,
                    Hostname:   m.Container.Config.Hostname,
                    Labels:     m.Container.Config.Labels,
                },
            },
            Level:      adapter.getLevel(m.Source),
            Hostname:   adapter.getHost(m.Container.Config.Hostname),
            Tags:       adapter.getTags(m),
        })

        if err != nil {
            log.Println(
                fmt.Errorf(
                    "JSON Marshalling Error: %s",
                    err.Error(),
                ),
            )
        } else {
            adapter.Queue <- Line{
                Line:       string(messageStr),
                File:       m.Container.Name,
                Timestamp:  time.Now().Unix(),
            }
        }
    }
}

// readQueue is a method for reading from queue:
func (adapter *Adapter) readQueue() {

    buffer := make([]Line, 0)
    timeout := time.NewTimer(adapter.Config.FlushInterval)

    for {
        select {
        case msg := <-adapter.Queue:
            if binary.Size(buffer) >= int(adapter.Config.MaxBufferSize) {
                timeout.Stop()
                adapter.flushBuffer(buffer)
                buffer = make([]Line, 0)
            }

            buffer = append(buffer, msg)

        case <-timeout.C:
            if len(buffer) > 0 {
                adapter.flushBuffer(buffer)
                buffer = make([]Line, 0)
            }
        }

        timeout.Reset(adapter.Config.FlushInterval)
    }
}

// flushBuffer is a method for flushing the lines:
func (adapter *Adapter) flushBuffer(buffer []Line) {
    var data bytes.Buffer

    
    body := struct {
        Lines []Line `json:"lines"`
    }{
        Lines: buffer,
    }


    if error := json.NewEncoder(&data).Encode(body); error != nil {
        log.Println(
            fmt.Errorf(
                "JSON Encoding Error: %s",
                error.Error(),
            ),
        )
        return
    }

    req, _ := http.NewRequest(http.MethodPost, buildLogDNAURL(adapter.Config.LogDNAURL, adapter.Config.LogDNAKey), &data)
    req.Header.Set("user-agent", "logspout/" + os.Getenv("BUILD_VERSION"))
    req.Header.Set("Content-Type", "application/json; charset=UTF-8")
    req.SetBasicAuth(adapter.Config.LogDNAKey, "")
    resp, err := adapter.HTTPClient.Do(req)

    if err != nil {
        log.Println(
            fmt.Errorf(
                "HTTP Client Post Request Error: %s",
                err.Error(),
            ),
        )
        return
    }

    if resp != nil {
        if resp.StatusCode != http.StatusOK {
            log.Println(
                fmt.Errorf(
                    "Received Status Code: %s While Sending Message",
                    resp.StatusCode,
                ),
            )
        }
        defer resp.Body.Close()
    }
}

func buildLogDNAURL(baseURL, token string) string {

    v := url.Values{}
    v.Add("apikey", token)
    v.Add("hostname", "logdna_logspout")

    return "https://" + baseURL + "?" + v.Encode()
}