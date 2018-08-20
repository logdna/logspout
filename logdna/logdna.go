package logdna

import (
    "errors"
    "strings"
    "log"
    "os"

    "github.com/gliderlabs/logspout/router"
    "github.com/smusali/logspout/logdna/adapter"
)

const (
    adapter             = "logdna"
    endpointVar         = "LOGDNA_URL"
    tokenVar            = "LOGDNA_KEY"
    tagsVar             = "TAGS"
    filterIDVar         = "FILTER_ID"
    filterNameVar       = "FILTER_NAME"
    filterNamesVar      = "FILTER_NAMES"
    filterSourcesVar    = "FILTER_SOURCES"
    filterLabelsVar     = "FILTER_LABELS"
)

func init() {
    router.AdapterFactories.Register(NewLogDNAAdapter, adapter)

    filterID := os.Getenv(filterIDVar)
    filterName := os.Getenv(filterNameVar)
    filterNames := strings.Split(os.Getenv(filterNamesVar), ",")
    filterSources := strings.Split(os.Getenv(filterSourcesVar), ",")
    filterLabels := strings.Split(os.Getenv(filterLabelsVar), ",")

    r := &router.Route{
        Adapter:        adapter,
        FilterID:       filterID,
        FilterName:     filterName,
        FilterNames:    filterNames,
        FilterSources:  filterSources,
        FilterLabels:   filterLabels,
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
    ), nil
}
