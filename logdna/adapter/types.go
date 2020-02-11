// Package adapter is the implementation of LogDNA LogSpout Adapter
package adapter

import (
    "time"
    "os"
    "strconv"
    "sync"

    "github.com/gojektech/heimdall"
)

/*
    Definitions of All Structs
*/

// Configuration is Configuration Struct for LogDNA Adapter:
type Configuration struct {
    FlushInterval   time.Duration
    Hostname        string
    LogDNAKey       string
    LogDNAURL       string
    MaxBufferSize   uint64
    Tags            string
}

// Adapter structure:
type Adapter struct {
    Config      Configuration
    Queue       chan Line
    HTTPClient  heimdall.Client
    sync.Mutex
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
    Config  ContainerConfig `json:"config"`
}

// ContainerConfig structure for the Config of ContainerInfo:
type ContainerConfig struct {
    Image       string              `json:"image"`
    Hostname    string              `json:"hostname"`
    Labels      map[string]string   `json:"labels"`
}

/*
    Common Functions
*/

// Getting Uint Variable from Environment:
func getUintOpt(name string, dfault uint64) uint64 {
    if result, err := strconv.ParseUint(os.Getenv(name), 10, 64); err == nil {
        return result
    }
    return dfault
}

// Getting Duration Variable from Environment:
func getDurationOpt(name string, dfault time.Duration) time.Duration {
    if result, err := strconv.ParseInt(os.Getenv(name), 10, 64); err == nil {
        return time.Duration(result)
    }
    return dfault
}

// Getting String Variable from Environment:
func getStringOpt(name, dfault string) string {
    if value := os.Getenv(name); value != "" {
        return value
    }
    return dfault
}