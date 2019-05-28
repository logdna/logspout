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
    "strings"
    "time"

    "github.com/gliderlabs/logspout/router"
    "github.com/answerbook/logspout/logdna/adapter"
)

func init() {
    router.AdapterFactories.Register(NewLogDNAAdapter, "logdna")

    r := &router.Route{
        Adapter:        "logdna",
        FilterName:     os.Getenv("FILTER_NAME"),
        FilterID:       os.Getenv("FILTER_ID"),
        FilterLabels:   strings.Split(os.Getenv("FILTER_LABELS"), ","),
        FilterSources:  strings.Split(os.Getenv("FILTER_SOURCES"), ","),
    }

    if err := router.Routes.Add(r); err != nil {
        log.Fatal("could not add route: ", err.Error())
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
            Endpoint:   "logs.logdna.com/logs/ingest",
            Hostname:   os.Getenv("HOSTNAME"),
            Tags:       strings.Split(os.Getenv("TAGS"), ","),
            Token:      token,
            Verbose:    true,
        }, HTTPClient:  adapter.HTTPClientConfiguration{
            DialContextKeepAlive:   60 * time.Second,   // 30 by Default
            DialContextTimeout:     60 * time.Second,   // 30 by Default
            ExpectContinueTimeout:  5 * time.Second,    // 1 by Default
            IdleConnTimeout:        60 * time.Second,   // 90 by Default
            Timeout:                60 * time.Second,   // 30 by Default
            TLSHandshakeTimeout:    30 * time.Second,   // 10 by Default
        }, Limits:      adapter.LimitConfiguration{
            FlushInterval:      250 * time.Millisecond,
            MaxBufferSize:      2 * 1024 * 1024,
            MaxLineLength:      16000,
            MaxRequestRetry:    10,
        },
    }

    endpoint := os.Getenv("LOGDNA_URL")
    if endpoint != "" {
        config.Custom.Endpoint = endpoint
    }

    if (os.Getenv("VERBOSE") == "0") {
        config.Custom.Verbose = false
    }

    // os.Setenv("INACTIVITY_TIMEOUT", "1m")

    return adapter.New(config), nil
}
