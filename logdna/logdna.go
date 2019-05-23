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

    "github.com/gliderlabs/logspout/router"
    "github.com/logdna/logspout/logdna/adapter"
)

const (
    endpointVar         = "LOGDNA_URL"
    tokenVar            = "LOGDNA_KEY"
    tagsVar             = "TAGS"
    hostVar             = "HOSTNAME"
    filterNameVar       = "FILTER_NAME"
    filterIDVar         = "FILTER_ID"
    filterSourcesVar    = "FILTER_SOURCES"
    filterLabelsVar     = "FILTER_LABELS"
)

func init() {
    router.AdapterFactories.Register(NewLogDNAAdapter, "logdna")

    filterLabels := make([]string, 0)
    filterLabelsValue := os.Getenv(filterLabelsVar)
    if filterLabelsValue != "" {
        filterLabels = strings.Split(filterLabelsValue, ",")
    }

    filterSources := make([]string, 0)
    filterSourcesValue := os.Getenv(filterSourcesVar)
    if filterSourcesValue != "" {
        filterSources = strings.Split(filterSourcesValue, ",")
    }

    filterID := os.Getenv(filterIDVar)
    filterName := os.Getenv(filterNameVar)

    r := &router.Route{
        Adapter:        "logdna",
        FilterName:     filterName,
        FilterID:       filterID,
        FilterLabels:   filterLabels,
        FilterSources:  filterSources,
    }

    err := router.Routes.Add(r)
    if err != nil {
        log.Fatal("could not add route: ", err.Error())
    }
}

// NewLogDNAAdapter creates adapter:
func NewLogDNAAdapter(route *router.Route) (router.LogAdapter, error) {
    endpoint := os.Getenv(endpointVar)
    token := os.Getenv(tokenVar)
    tags := os.Getenv(tagsVar)
    hostname := os.Getenv(hostVar)

    if endpoint == "" {
        endpoint = "logs.logdna.com/logs/ingest"
    }

    if token == "" {
        return nil, errors.New(
            "could not find environment variable LOGDNA_KEY",
        )
    }

    if hostname == "" {
        hostname = "no_custom_hostname"
    }

    return adapter.New(
        endpoint,
        token,
        tags,
        hostname,
    ), nil
}
