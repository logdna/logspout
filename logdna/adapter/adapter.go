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
    log.Println("I'm in Adapter")
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
    log.Println("Calling readQueue")
    go adapter.readQueue()
    return adapter
}

func (adapter *Adapter) getLevel(source string) string {
    switch source {
    case "stdout":
        return "INFO"
    case "stderr":
        return "ERROR"
    }
    return ""
}

func (adapter *Adapter) getHost(containerHostname string) string {
    host := containerHostname
    if (adapter.Config.Hostname != "") {
        host = adapter.Config.Hostname
    }
    return host
}

func (adapter *Adapter) getTags(m *router.Message) string {
    
    var listTags []string
    existenceMap := map[string]bool{}

    for _, t := range adapter.Config.Tags {
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

func (adapter *Adapter) sanitizeMessage(message string) string {
    if uint64(len(message)) > adapter.Limits.MaxLineLength {
        return message[0:adapter.Limits.MaxLineLength] + " (cut off, too long...)"
    }
    return message
}

// Stream method is for streaming the messages:
func (adapter *Adapter) Stream(logstream chan *router.Message) {
    log.Print("Log Stream: ", logstream)
    for m := range logstream {
        log.Print("Message: ", m.Data)
//        if adapter.Config.Verbose || m.Container.Config.Image != "logdna/logspout" {
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
//        }
    }
}

func (adapter *Adapter) readQueue() {

    log.Println("I am in readQueue")
    buffer := make([]Line, 0)
    timeout := time.NewTimer(adapter.Limits.FlushInterval)
    bytes := 0

    for {
        select {
        case msg := <-adapter.Queue:
            if uint64(bytes) >= adapter.Limits.MaxBufferSize {
                timeout.Stop()
                adapter.flushBuffer(buffer)
                timeout.Reset(adapter.Limits.FlushInterval)
                buffer = make([]Line, 0)
                bytes = 0
            }

            buffer = append(buffer, msg)
            bytes += len(msg.Line)

        case <-timeout.C:
            if len(buffer) > 0 {
                adapter.flushBuffer(buffer)
                timeout.Reset(adapter.Limits.FlushInterval)
                buffer = make([]Line, 0)
                bytes = 0
            }
        }
    }
}

func (adapter *Adapter) flushBuffer(buffer []Line) {
    var data bytes.Buffer

    body := struct {
        Lines []Line `json:"lines"`
    }{
        Lines: buffer,
    }

    err := json.NewEncoder(&data).Encode(body)

    if err != nil {
        adapter.Log.Println(
            fmt.Errorf(
                "JSON Encoding Error: %s",
                err.Error(),
            ),
        )
        return
    }

    resp, err := adapter.HTTPClient.Post(adapter.LogDNAURL, "application/json; charset=UTF-8", &data)

    if resp != nil {
        defer resp.Body.Close()
    }

    if err != nil {
        if _, ok := err.(net.Error); ok {
            for _, line := range buffer {
                if line.Retried < adapter.Limits.MaxRequestRetry {
                    line.Retried++
                    adapter.Queue <- line
                }
            }
        } else {
            adapter.Log.Println(
                fmt.Errorf(
                    "HTTP Client Post Request Error: %s",
                    err.Error(),
                ),
            )
        }
        return
    }

    if resp.StatusCode != http.StatusOK {
        adapter.Log.Println(
            fmt.Errorf(
                "Received Status Code: %s While Sending Message.\nResponse: %s",
                resp.StatusCode,
                resp.Body,
            ),
        )
    }
}

func buildLogDNAURL(baseURL, token string) string {

    v := url.Values{}
    v.Add("apikey", token)
    v.Add("hostname", "logdna_logspout")

    ldnaURL := "https://" + baseURL + "?" + v.Encode()
    return ldnaURL
}

/*
    Definitions of All Structs
*/

// Configuration is Configuration Struct for LogDNA Adapter:
type Configuration struct {
    Custom      CustomConfiguration
    HTTPClient  HTTPClientConfiguration
    Limits      LimitConfiguration
}

// CustomConfiguration is Custom SubConfiguration:
type CustomConfiguration struct {
    Endpoint    string
    Hostname    string
    Tags        []string
    Token       string
    Verbose     bool
}

// LimitConfiguration is SubConfiguration for Limits:
type LimitConfiguration struct {
    FlushInterval   time.Duration
    MaxBufferSize   uint64
    MaxLineLength   uint64
    MaxRequestRetry uint64
}

// HTTPClientConfiguration is for Configuring HTTP Client:
type HTTPClientConfiguration struct {
    DialContextTimeout      time.Duration
    DialContextKeepAlive    time.Duration
    ExpectContinueTimeout   time.Duration
    IdleConnTimeout         time.Duration
    Timeout                 time.Duration
    TLSHandshakeTimeout     time.Duration
}

// Adapter structure:
type Adapter struct {
    Config      CustomConfiguration
    Limits      LimitConfiguration
    Log         *log.Logger
    LogDNAURL   string
    Queue       chan Line
    HTTPClient  *http.Client
}

// Line structure for the queue of Adapter:
type Line struct {
    Timestamp   int64  `json:"timestamp"`
    Line        string `json:"line"`
    File        string `json:"file"`
    Retried     uint64 `json:"-"`
}

// Message structure:
type Message struct {
    Message     string        `json:"message"`
    Container   ContainerInfo `json:"container"`
    Level       string        `json:"level"`
    Hostname    string        `json:"hostname"`
    Tags        string        `json:"tags"`
}

// ContainerInfo structure for the Container of Message:
type ContainerInfo struct {
    Name    string          `json:"name"`
    ID      string          `json:"id"`
    PID     int             `json:"pid",omitempty`
    Config  ContainerConfig `json:"config"`
}

// ContainerConfig structure for the Config of ContainerInfo:
type ContainerConfig struct {
    Image       string              `json:"image"`
    Hostname    string              `json:"hostname"`
    Labels      map[string]string   `json:"labels"`
}