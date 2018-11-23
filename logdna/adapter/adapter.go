package adapter

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/url"
    "net/http"
    "os"
    "time"
    "strings"
    "text/template"

    "github.com/gliderlabs/logspout/router"
)

const (
    flushTimeout    = time.Second
    bufferSize      = 10000
)

// Adapter structure:
type Adapter struct {
    log        *log.Logger
    logdnaURL  string
    queue      chan Line
    host       string
    tags       string
}

// Line structure for the queue of Adapter:
type Line struct {
    Timestamp int64  `json:"timestamp"`
    Line      string `json:"line"`
    File      string `json:"file"`
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
    Config  ContainerConfig `json:"config"`
}

// ContainerConfig structure for the Config of ContainerInfo:
type ContainerConfig struct {
    Image       string              `json:"image"`
    Hostname    string              `json:"hostname"`
    Labels      map[string]string   `json:"labels"`
}

// New method of Adapter:
func New(baseURL string, logdnaToken string, tags string, hostname string) *Adapter {
    adapter := &Adapter{
        log:        log.New(os.Stdout, "logdna/logspout", log.LstdFlags),
        logdnaURL:  buildLogDNAURL(baseURL, logdnaToken),
        queue:      make(chan Line),
        host:       hostname,
        tags:       tags,
    }
    go adapter.readQueue()
    return adapter
}

func (adapter *Adapter) getLevel(source string) string {
    level := "ERROR"
    if (source == "stdout") {
        level = "INFO"
    }
    return level
}

func (adapter *Adapter) getHost(containerHostname string) string {
    host := containerHostname
    if (adapter.host != "no_custom_hostname") {
        host = adapter.host
    }
    return host
}

func (adapter *Adapter) getTags(m *router.Message) string {
    
    if adapter.tags == "" {
        adapter.log.Println("Empty %s", adapter.tags)
        return ""
    }

    adapter.log.Println(adapter.tags)
    fmt.Println(adapter.tags)

    splitTags := strings.Split(adapter.tags, ",")
    var listTags []string
    existenceMap := map[string]bool{}

    for _, t := range splitTags {
        if (strings.Contains(t, "{{") || strings.Contains(t, "}}")) {
            var parsedTagBytes bytes.Buffer
            
            adapter.log.Println(t)
            fmt.Println(t)

            tmp, e := template.New("parsedTag").Parse(t)
            if e == nil {
                err := tmp.ExecuteTemplate(&parsedTagBytes, "parsedTag", m)
                if err == nil {
                    parsedTag := parsedTagBytes.String()
                    for _, p := range strings.Split(parsedTag, ":") {
                        if existenceMap[p] == false {
                            listTags = append(listTags, p)
                            existenceMap[p] = true
                        }
                    }
                } else {
                    adapter.log.Println(
                        fmt.Errorf(
                            "Invalid Tag: %s",
                            t,
                        ),
                    )
                }
            } else {
                adapter.log.Println(
                    fmt.Errorf(
                        "Error in creating Template from %s",
                        t,
                    ),
                )
            }
        } else {
            if existenceMap[t] == false {
                listTags = append(listTags, t)
                existenceMap[t] = true
            }
        }
    }

    return strings.Join(listTags, ",")
}

// Stream method is for streaming the messages:
func (adapter *Adapter) Stream(logstream chan *router.Message) {
    for m := range logstream {
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

        adapter.log.Println(adapter.tags)
        fmt.Println(adapter.tags)

        if err != nil {
            adapter.log.Println(
                fmt.Errorf(
                    "Invalid Data: %s",
                    m.Data,
                ),
            )
        } else {
            adapter.queue <- Line{
                Line:       string(messageStr),
                File:       m.Container.Name,
                Timestamp:  time.Now().Unix(),
            }
        }
    }
}

func (adapter *Adapter) readQueue() {

    buffer := adapter.newBuffer()
    timeout := time.NewTimer(flushTimeout)
    for {
        select {
        case msg := <-adapter.queue:
            if len(buffer) == cap(buffer) {
                timeout.Stop()
                adapter.flushBuffer(buffer)
                buffer = adapter.newBuffer()
            }

            buffer = append(buffer, msg)

        case <-timeout.C:
            if len(buffer) > 0 {
                adapter.flushBuffer(buffer)
                buffer = adapter.newBuffer()
            }
        }

        timeout.Reset(flushTimeout)
    }
}

func (adapter *Adapter) newBuffer() []Line {
    return make([]Line, 0, bufferSize)
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
        adapter.log.Println(
            fmt.Errorf(
                "error from client: %s",
                "following lines couldn't be encoded:",
            ),
        )
        for i, line := range buffer {
            adapter.log.Println(
                fmt.Errorf(
                    "%d. %s",
                    i,
                    line.Line,
                ),
            )
        }
        return
    }

    resp, err := http.Post(adapter.logdnaURL, "application/json; charset=UTF-8", &data)

    if resp != nil {
        defer resp.Body.Close()
    }

    if err != nil {
        adapter.log.Println(
            fmt.Errorf(
                "error from client: %s",
                err.Error(),
            ),
        )
        return
    }

    if resp.StatusCode != http.StatusOK {
        adapter.log.Println(
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
