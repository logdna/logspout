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
    "github.com/smusali/logspout/logdna/adapter"
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
    logdnaKey := os.Getenv("LOGDNA_KEY")
    if logdnaKey == "" {
        return nil, errors.New("Cannot Find Environment Variable \"LOGDNA_KEY\"")
    }

    hostname := os.Getenv("HOSTNAME")
    if hostname == "" {
        hostname, _ := os.Hostname()
    }

    if os.Getenv("INACTIVITY_TIMEOUT") == "" {
        os.Setenv("INACTIVITY_TIMEOUT", "1m")
    }

    config := adapter.Configuration{
        FlushInterval:  getDurationOpt("FLUSH_INTERVAL", 250) * time.Millisecond,
        Hostname:       hostname,
        LogDNAKey:      logdnaKey,
        LogDNAURL:      getStringOpt("LOGDNA_URL", "logs.logdna.com/logs/ingest"),
        MaxBufferSize:  getUintOpt("MAX_BUFFER_SIZE", 2) * 1024 * 1024,
        Tags:           os.Getenv("TAGS"),
    }

    return adapter.New(config), nil
}