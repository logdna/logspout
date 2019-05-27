// Package adapter is the implementation of LogDNA LogSpout Adapter
package adapter

import (
    "log"
    "net/http"
    "time"
)

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