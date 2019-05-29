// Package logdna is LogSpout Adapter to Forward Logs to LogDNA 
package logdna

// The MIT License (MIT)
// =====================
//
// Copyright (c) 2019 LogDNA, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
    "errors"
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/gliderlabs/logspout/router"
    "github.com/logdna/logspout/logdna/adapter"
)

func init() {
    router.AdapterFactories.Register(NewLogDNAAdapter, "logdna")

    filterLabels := make([]string, 0)
    if filterLabelsValue := os.Getenv("FILTER_LABELS"); filterLabelsValue != "" {
        filterLabels = strings.Split(filterLabelsValue, ",")
    }

    filterSources := make([]string, 0)
    if filterSourcesValue := os.Getenv("FILTER_SOURCES"); filterSourcesValue != "" {
        filterSources = strings.Split(filterSourcesValue, ",")
    }

    r := &router.Route{
        Adapter:        "logdna",
        FilterName:     getStringOpt("FILTER_NAME", ""),
        FilterID:       getStringOpt("FILTER_ID", ""),
        FilterLabels:   filterLabels,
        FilterSources:  filterSources,
    }

    if err := router.Routes.Add(r); err != nil {
        log.Fatal("Cannot Add New Route: ", err.Error())
    }
}

// NewLogDNAAdapter creates adapter:
func NewLogDNAAdapter(route *router.Route) (router.LogAdapter, error) {
    token := os.Getenv("LOGDNA_KEY")
    if token == "" {
        return nil, errors.New("Cannot Find Environment Variable \"LOGDNA_KEY\"")
    }

    config := adapter.Configuration{
        Custom:         adapter.CustomConfiguration{
            Endpoint:   getStringOpt("LOGDNA_URL", "logs.logdna.com/logs/ingest"),
            Hostname:   getStringOpt("HOSTNAME", ""),
            Tags:       getStringOpt("TAGS", ""),
            Token:      token,
            Verbose:    os.Getenv("VERBOSE") != "0",
        }, HTTPClient:  adapter.HTTPClientConfiguration{
            DialContextKeepAlive:   getDurationOpt("DIAL_KEEP_ALIVE", 60) * time.Second,        // 30 by Default
            DialContextTimeout:     getDurationOpt("DIAL_TIMEOUT", 60) * time.Second,           // 30 by Default
            ExpectContinueTimeout:  getDurationOpt("EXPECT_CONTINUE_TIMEOUT", 5) * time.Second, // 1 by Default
            IdleConnTimeout:        getDurationOpt("IDLE_CONN_TIMEOUT", 60) * time.Second,      // 90 by Default
            Timeout:                getDurationOpt("HTTP_CLIENT_TIMEOUT", 60) * time.Second,    // 30 by Default
            TLSHandshakeTimeout:    getDurationOpt("TLS_HANDSHAKE_TIMEOUT", 30) * time.Second,   // 10 by Default
        }, Limits:      adapter.LimitConfiguration{
            FlushInterval:      getDurationOpt("FLUSH_INTERVAL", 250) * time.Millisecond,
            MaxBufferSize:      getUintOpt("MAX_BUFFER_SIZE", 2) * 1024 * 1024,
            MaxLineLength:      getUintOpt("MAX_LINE_LENGTH", 16000),
            MaxRequestRetry:    getUintOpt("MAX_REQUEST_RETRY", 10),
        },
    }

    if os.Getenv("INACTIVITY_TIMEOUT") == "" {
        os.Setenv("INACTIVITY_TIMEOUT", "1m")
    }

    return adapter.New(config), nil
}

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