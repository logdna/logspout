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

    filter_labels := make([]string, 0)
    filterLabelsValue := os.Getenv(filterLabelsVar)
    if filterLabelsValue != "" {
        filter_labels = strings.Split(filterLabelsValue, ",")
    }

    filter_sources := make([]string, 0)
    filterSourcesValue := os.Getenv(filterSourcesVar)
    if filterSourcesValue != "" {
        filter_sources = strings.Split(filterSourcesValue, ",")
    }

    filter_id := os.Getenv(filterIDVar)
    filter_name := os.Getenv(filterNameVar)

    r := &router.Route{
        Adapter:        "logdna",
        FilterName:     filter_name,
        FilterID:       filter_id,
        FilterLabels:   filter_labels,
        FilterSources:  filter_sources,
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
