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
    "github.com/logdna/logspout/logdna/types"
)

// New method of Adapter:
func New(config types.Configuration) *types.Adapter {
    adapter := &types.Adapter{
        Config:     config.Custom,
        Limits:     config.Limits,
        Log:        log.New(os.Stdout, config.Custom.Hostname + " ", log.LstdFlags),
        LogDNAURL:  buildLogDNAURL(config.Custom.Endpoint, config.Custom.Token),
        Queue:      make(chan types.Line),
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

func (adapter *types.Adapter) getLevel(source string) string {
    switch source {
    case "stdout":
        return "INFO"
    case "stderr":
        return "ERROR"
    }
    return ""
}

func (adapter *types.Adapter) getHost(containerHostname string) string {
    host := containerHostname
    if (adapter.Config.Hostname != "") {
        host = adapter.Config.Hostname
    }
    return host
}

func (adapter *types.Adapter) getTags(m *router.Message) string {
    
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

func (adapter *types.Adapter) sanitizeMessage(message string) string {
    if uint64(len(message)) > adapter.Limits.MaxLineLength {
        return message[0:adapter.Limits.MaxLineLength] + " (cut off, too long...)"
    }
    return message
}

// Stream method is for streaming the messages:
func (adapter *types.Adapter) Stream(logstream chan *router.Message) {
    for m := range logstream {
        if adapter.Config.Verbose || m.Container.Config.Image != "logdna/logspout" {
            messageStr, err := json.Marshal(Message{
                Message:    adapter.sanitizeMessage(m.Data),
                Container:  ContainerInfo{
                    Name:   m.Container.Name,
                    ID:     m.Container.ID,
                    Config: ContainerConfig{
                        Image:      m.Container.Config.Image,
                        Hostname:   m.Container.Config.Custom.Hostname,
                        Labels:     m.Container.Config.Labels,
                    },
                },
                Level:      adapter.getLevel(m.Source),
                Hostname:   adapter.getHost(m.Container.Config.Custom.Hostname),
                Tags:       adapter.getTags(m),
            })

            if err != nil {
                adapter.Log.Println(
                    fmt.Errorf(
                        "Invalid Data: %s",
                        m.Data,
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
}

func (adapter *types.Adapter) readQueue() {

    buffer := make([]Line, 0)
    timeout := time.NewTimer(adapter.Limits.FlushInterval)
    bytes := 0

    for {
        select {
        case msg := <-adapter.Queue:
            if bytes >= adapter.Limits.MaxBufferSize {
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

func (adapter *types.Adapter) flushBuffer(buffer []Line) {
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
                "error from client: %s",
                "following lines couldn't be encoded:",
            ),
        )
        for i, line := range buffer {
            adapter.Log.Println(
                fmt.Errorf(
                    "%d. %s",
                    i,
                    line.Line,
                ),
            )
        }
        return
    }

    resp, err := adapter.HTTPClient.Post(adapter.LogDNAURL, "application/json; charset=UTF-8", &data)

    if resp != nil {
        defer resp.Body.Close()
    }

    if err != nil {
        if err, ok := err.(net.Error); ok && err.Timeout() {
            for _, line := range buffer {
                adapter.Queue <- line
            }
        } else {
            adapter.Log.Println(
                fmt.Errorf(
                    "error from client: %s",
                    err.Error(),
                ),
            )
        }
        return
    }

    if resp.StatusCode != http.StatusOK {
        adapter.Log.Println(
            fmt.Errorf(
                "received a %s status code when sending message. response: %s",
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
