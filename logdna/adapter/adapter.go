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

func New(baseURL string, logdnaToken string, tags string) *Adapter {
    adapter := &Adapter{
        log:        log.New(os.Stdout, "logspout-logdna", log.LstdFlags),
        logdnaURL:  buildLogDNAURL(baseURL, logdnaToken, tags),
        queue:      make(chan Line),
    }

    log.Println(&adapter)
    go adapter.readQueue()
    log.Println(&adapter)
    return adapter
}

func (l *Adapter) getLevel(source string) string {
    level := "ERROR"
    if (source == "stdout") {
        level = "INFO"
    }
    log.Println(level)
    return level
}

func (l *Adapter) Stream(logstream chan *router.Message) {
    log.Println(&logstream)
    for m := range logstream {
        log.Println(m)
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
            Level:      l.getLevel(m.Source),
        })
        log.Println(messageStr)
        if err != nil {
            log.Fatal(err.Error())
        }
        l.queue <- Line{
            Line:       string(messageStr),
            File:       m.Container.Name,
            Timestamp:  time.Now().Unix(),
        }
    }
}

func (l *Adapter) readQueue() {

    buffer := l.newBuffer()
    timeout := time.NewTimer(flushTimeout)
    log.Println(&buffer)
    log.Println(&timeout)
    for {
        select {
        case msg := <-l.queue:
            if len(buffer) == cap(buffer) {
                timeout.Stop()
                l.flushBuffer(buffer)
                buffer = l.newBuffer()
            }

            buffer = append(buffer, msg)

        case <-timeout.C:
            if len(buffer) > 0 {
                l.flushBuffer(buffer)
                buffer = l.newBuffer()
            }
        }

        timeout.Reset(flushTimeout)
    }
}

func (l *Adapter) newBuffer() []Line {
    return make([]Line, 0, bufferSize)
}

func (l *Adapter) flushBuffer(buffer []Line) {
    var data bytes.Buffer

    body := struct {
        Lines []Line `json:"lines"`
    }{
        Lines: buffer,
    }

    log.Println(body)

    json.NewEncoder(&data).Encode(body)
    resp, err := http.Post(l.logdnaURL, "application/json; charset=UTF-8", &data)

    log.Println(resp)
    log.Println(resp.StatusCode)

    if resp != nil {
        defer resp.Body.Close()
    }

    if err != nil {
        l.log.Println(
            fmt.Errorf(
                "error from client: %s",
                err.Error(),
            ),
        )
        return
    }

    if resp.StatusCode != http.StatusOK {
        l.log.Println(
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
    v.Add("tags", tags)
    v.Add("apikey", token)

    ldna_url := "https://" + baseURL + "?" + v.Encode()

    return ldna_url
}
