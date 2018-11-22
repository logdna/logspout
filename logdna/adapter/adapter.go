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

    "github.com/gliderlabs/logspout/router"
)

const (
    flushTimeout    = time.Second
    bufferSize      = 10000
)

type Adapter struct {
    log        *log.Logger
    logdnaURL  string
    queue      chan Line
    host       string
}

type Line struct {
    Timestamp int64  `json:"timestamp"`
    Line      string `json:"line"`
    File      string `json:"file"`
}

type Message struct {
    Message     string        `json:"message"`
    Container   ContainerInfo `json:"container"`
    Level       string        `json:"level"`
    Hostname    string        `json:"hostname"`
}

type ContainerInfo struct {
    Name    string          `json:"name"`
    ID      string          `json:"id"`
    Config  ContainerConfig `json:"config"`
}

type ContainerConfig struct {
    Image       string              `json:"image"`
    Hostname    string              `json:"hostname"`
    Labels      map[string]string   `json:"labels"`
}

func New(baseURL string, logdnaToken string, tags string, hostname string) *Adapter {
    adapter := &Adapter{
        log:        log.New(os.Stdout, "logspout-logdna", log.LstdFlags),
        logdnaURL:  buildLogDNAURL(baseURL, logdnaToken, tags),
        queue:      make(chan Line),
        host:       hostname,
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

func (adapter *Adapter) getHost(container_hostname string) string {
    host := container_hostname
    if (adapter.host != "no_custom_hostname") {
        host = adapter.host
    }
    return host
}

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
        })
        if err != nil {
            log.Fatal(err.Error())
        }
        adapter.queue <- Line{
            Line:       string(messageStr),
            File:       m.Container.Name,
            Timestamp:  time.Now().Unix(),
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

    json.NewEncoder(&data).Encode(body)
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

func buildLogDNAURL(baseURL, token string, tags string) string {

    v := url.Values{}
    if tags != "" {
        v.Add("tags", tags)
    }
    v.Add("apikey", token)
    v.Add("hostname", "logdna_logspout")

    ldna_url := "https://" + baseURL + "?" + v.Encode()
    return ldna_url
}
