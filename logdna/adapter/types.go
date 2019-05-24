// Package types defines the types used in LogDNA LogSpout Adapter
package types

import (
    "log"
)

// Configuration is Configuration Struct for LogDNA Adapter:
type Configuration struct {
    Endpoint        string
    FlushInterval   uint64
    Hostname        string
    MaxBufferSize   uint64
    Tags            []string
    Token           string
    Verbose			bool
}

// Adapter structure:
type Adapter struct {
    Log        *log.Logger
    LogdnaURL  string
    Queue      chan Line
    Config     Configuration
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