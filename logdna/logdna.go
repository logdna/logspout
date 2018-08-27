package logdna

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

    r := &router.Route{
        Adapter:        "logdna",
        FilterName:     os.Getenv(filterNameVar),
        FilterID:       os.Getenv(filterIDVar),
        FilterLabels:   strings.Split(os.Getenv(filterLabelsVar), ","),
        FilterSources:  strings.Split(os.Getenv(filterSourcesVar), ","),
    }

    err := router.Routes.Add(r)
    if err != nil {
        log.Fatal("could not add route: ", err.Error())
    }
}

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

    return adapter.New(
        endpoint,
        token,
        tags,
        hostname,
    ), nil
}
