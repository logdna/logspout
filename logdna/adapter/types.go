// Package adapter is the implementation of LogDNA LogSpout Adapter
package adapter

import (
	"log"
	"time"

	"github.com/gojektech/heimdall"
)

/*
   Definitions of All Structs
*/

// Configuration is Configuration Struct for LogDNA Adapter:
type Configuration struct {
	BackoffInterval   time.Duration
	FlushInterval     time.Duration
	Hostname          string
	HTTPTimeout       time.Duration
	JitterInterval    time.Duration
	LogDNAKey         string
	LogDNAURL         string
	MaxBufferSize     uint64
	RequestRetryCount uint64
	Tags              string
}

// Adapter structure:
type Adapter struct {
	Config        Configuration
	HTTPClient    heimdall.Client
	Logger        *log.Logger
	Queue         chan Line
	instrumenting *instrumentingAdapter
	version       string
}

// Line structure for the queue of Adapter:
type Line struct {
	Timestamp int64  `json:"timestamp"`
	Line      string `json:"line"`
	File      string `json:"file"`
}

// Message structure:
type Message struct {
	Message   string        `json:"message"`
	Container ContainerInfo `json:"container"`
	Level     string        `json:"level"`
	Hostname  string        `json:"hostname"`
	Tags      string        `json:"tags"`
}

// ContainerInfo structure for the Container of Message:
type ContainerInfo struct {
	Name   string          `json:"name"`
	ID     string          `json:"id"`
	Config ContainerConfig `json:"config"`
}

// ContainerConfig structure for the Config of ContainerInfo:
type ContainerConfig struct {
	Image    string            `json:"image"`
	Hostname string            `json:"hostname"`
	Labels   map[string]string `json:"labels"`
}
