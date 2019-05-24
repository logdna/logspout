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

    "github.com/gliderlabs/logspout/router"
    "github.com/logdna/logspout/logdna/adapter"
    "github.com/logdna/logspout/logdna/types"
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

    config := types.Configuration{
        Endpoint:       "logs.logdna.com/logs/ingest",
        FlushInterval:  250,
        Hostname:       os.Getenv("HOSTNAME"),
        MaxBufferSize:  2 * 1024 * 1024,
        Token:          token,
        Tags:           strings.Split(os.Getenv("TAGS"), ","),
    }

    endpoint := os.Getenv("LOGDNA_URL")
    if endpoint != "" {
        config.Endpoint = endpoint
    }

    return adapter.New(config), nil
}
