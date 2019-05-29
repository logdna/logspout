// Package adapter is the implementation of LogDNA LogSpout Adapter
package adapter

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "strings"
    "text/template"
    "time"

    "github.com/gliderlabs/logspout/router"
)

// New method of Adapter:
func New(config Configuration) *Adapter {
    adapter := &Adapter{
        Config:     config.Custom,
        Limits:     config.Limits,
        Log:        log.New(os.Stdout, config.Custom.Hostname + " ", log.LstdFlags),
        LogDNAURL:  buildLogDNAURL(config.Custom.Endpoint, config.Custom.Token),
        Queue:      make(chan Line),
        HTTPClient: &http.Client{
            Timeout:    config.HTTPClient.Timeout,
            Transport:  &http.Transport{
                ExpectContinueTimeout:  config.HTTPClient.ExpectContinueTimeout,
                IdleConnTimeout:        config.HTTPClient.IdleConnTimeout,
                TLSHandshakeTimeout:    config.HTTPClient.TLSHandshakeTimeout,
                DialContext: (&net.Dialer{
                    KeepAlive: config.HTTPClient.DialContextKeepAlive,
                    Timeout:   config.HTTPClient.DialContextTimeout,
                }).DialContext,
            },
        },
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

// sanitizeMessage is a method for sanitizing the log message:
func (adapter *Adapter) sanitizeMessage(message string) string {
    if uint64(len(message)) > adapter.Limits.MaxLineLength {
        return message[0:adapter.Limits.MaxLineLength] + " (cut off, too long...)"
    }
    return message
}

// Stream method is for streaming the messages:
func (adapter *Adapter) Stream(logstream chan *router.Message) {
    for m := range logstream {
        if adapter.Config.Verbose || m.Container.Config.Image != "logdna/logspout" {
            messageStr, err := json.Marshal(Message{
                Message:    adapter.sanitizeMessage(m.Data),
                Container:  ContainerInfo{
                    Name:   m.Container.Name,
                    ID:     m.Container.ID,
                    PID:    m.Container.State.Pid,
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
                adapter.Log.Println(
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
                    Retried:    0,
                }
            }
        }
    }
}

// readQueue is a method for reading from queue:
func (adapter *Adapter) readQueue() {

    buffer := make([]Line, 0)
    timeout := time.NewTimer(adapter.Limits.FlushInterval)
    byteSize := 0

    for {
        select {
        case msg := <-adapter.Queue:
            if uint64(byteSize) >= adapter.Limits.MaxBufferSize {
                timeout.Stop()
                adapter.flushBuffer(buffer)
                timeout.Reset(adapter.Limits.FlushInterval)
                buffer = make([]Line, 0)
                byteSize = 0
            }

            buffer = append(buffer, msg)
            byteSize += len(msg.Line)

        case <-timeout.C:
            if len(buffer) > 0 {
                adapter.flushBuffer(buffer)
                timeout.Reset(adapter.Limits.FlushInterval)
                buffer = make([]Line, 0)
                byteSize = 0
            }
        }
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
        adapter.Log.Println(
            fmt.Errorf(
                "JSON Encoding Error: %s",
                error.Error(),
            ),
        )
        return
    }

    resp, err := adapter.HTTPClient.Post(adapter.LogDNAURL, "application/json; charset=UTF-8", &data)

    if resp != nil {
        adapter.Log.Print("Received Status Code: ", resp.StatusCode)
        defer resp.Body.Close()
    }

    if err != nil {
        if _, ok := err.(net.Error); ok {
            go adapter.retry(buffer)
        } else {
            adapter.Log.Println(
                fmt.Errorf(
                    "HTTP Client Post Request Error: %s",
                    err.Error(),
                ),
            )
        }
    }

    if resp.StatusCode != http.StatusOK {
        adapter.Log.Println(
            fmt.Errorf(
                "Received Status Code: %s While Sending Message",
                resp.StatusCode,
            ),
        )
    }
}

// retry sending the buffer:
func (adapter *Adapter) retry(buffer []Line) {
    adapter.Log.Print("Retrying...")
    for _, line := range buffer {
        if line.Retried < adapter.Limits.MaxRequestRetry {
            line.Retried++
            adapter.Queue <- line
        }
    }
}

func buildLogDNAURL(baseURL, token string) string {

    v := url.Values{}
    v.Add("apikey", token)
    v.Add("hostname", "logdna_logspout")

    return "https://" + baseURL + "?" + v.Encode()
}