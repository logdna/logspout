// Package types defines the types used in LogDNA LogSpout Adapter
package types

import (
	"context"
    "log"
    "net"
    "time"
)

/*
	Adapter{
		Log 		*log.Logger,
		LogDNAURL 	string,
		Config Configuration{
			Env Environment{
				Endpoint string,
				Hostname string,
				Tags 	 []string,
				Token	 string,
				Verbose	 bool,
			}, HTTP HTTPClientConfig{
				Timeout time.Duration,
				HTTPTransport HTTPTransportConfig{
					IdleConnTimeout 	time.Duration,
					MaxConnsPerHost 	int,
					MaxIdleConnsPerHost int,
					TLSHandshakeTimeout time.Duration,
					DialContext 		(&DialerConfig{
						Timeout: 	time.Duration,
						KeepAlive:	time.Duration,
					}).DialContext,
				},
			}, Ship ShippingConfig{
				FlushInterval time.Duration,
				MaxBufferSize uint64,
				MaxLineLength uint64,
			},
		},
		Queue chan Line{
			Timestamp int64,
			Line 	  string,
			File      string,
		},
	}
*/

// Adapter Structure
type Adapter struct {
    Log        *log.Logger
    LogDNAURL  string
    Config     Configuration
    Queue      chan Line
}

// Configuration is Configuration Struct for LogDNA Adapter
type Configuration struct {
    Env     Environment
    HTTP 	HTTPClientConfig
    Ship 	ShippingConfig
}

// Environment is the Struct to Store the Variables Configurable by Environment
type Environment struct {
	Endpoint	string
	Hostname    string
	Tags        []string
    Token       string
    Verbose		bool
}

// HTTPClientConfig is Configuration Struct for Creating HTTP Clients
type HTTPClientConfig struct {
	Timeout 		time.Duration
	HTTPTransport 	HTTPTransportConfig
}

// HTTPTransportConfig Structure
type HTTPTransportConfig struct {
	IdleConnTimeout		time.Duration
	MaxConnsPerHost		int
    MaxIdleConnsPerHost	int
	TLSHandshakeTimeout time.Duration
	DialContext 		func(ctx context.Context, network, addr string) (net.Conn, error)
}

// DialerConfig is a Structure for DialContext
type DialerConfig struct {
	Timeout 	time.Duration
	KeepAlive 	time.Duration
}

// ShippingConfig is the Struct to Store the Variables Related to Flushing and Buffering
type ShippingConfig struct {
	FlushInterval	time.Duration
	MaxBufferSize   uint64
    MaxLineLength	uint64
}

// Line Structure for the Queue of Adapter
type Line struct {
    Timestamp int64  `json:"timestamp"`
    Line      string `json:"line"`
    File      string `json:"file"`
}

/*
	Message{
		Hostname 	string,
		Level 		string,
		Message 	string,
		Tags 		string,
		Container   ContainerInfo{
			ID 		string,
			Name 	string,
			Config  ContainerConfig{
				Hostname 	string,
				Image 		string,
				Labels 		map[string]string,
			},
		},
	}
*/

// Message Structure
type Message struct {
    Hostname    string        `json:"hostname"`
    Level       string        `json:"level"`
    Message     string        `json:"message"`
    Tags        string        `json:"tags"`
    Container   ContainerInfo `json:"container"`
}

// ContainerInfo Structure for the Container of Message
type ContainerInfo struct {
	ID      string          `json:"id"`
    Name    string          `json:"name"`
    Config  ContainerConfig `json:"config"`
}

// ContainerConfig Structure for the Config of ContainerInfo
type ContainerConfig struct {
	Hostname    string              `json:"hostname"`
    Image       string              `json:"image"`
    Labels      map[string]string   `json:"labels"`
}