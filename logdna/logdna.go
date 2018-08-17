package logdna

import (
    "errors"
    "strings"
    "log"
    "os"

    "github.com/gliderlabs/logspout/router"
    "github.com/smusali/logspout/logdna/adapter"
)

func init() {
    router.AdapterFactories.Register(NewLogDNAAdapter, "logdna")

    r := &router.Route{
        Adapter:    "logdna",
    }

    err := router.Routes.Add(r)
    if err != nil {
        log.Fatal("could not add route: ", err.Error())
    }
}

func NewLogDNAAdapter(route *router.Route) (router.LogAdapter, error) {
    endpoint := os.Getenv("LOGDNA_URL")
    token := os.Getenv("LOGDNA_KEY")
    tags := os.Getenv("TAGS")
    hostname := os.Getenv("HOSTNAME")
    included := os.Getenv("INCLUDE")
    excluded := os.Getenv("EXCLUDE")

    if endpoint == "" {
        endpoint = "logs.logdna.com/logs/ingest"
    }

    if token == "" {
        return nil, errors.New(
            "could not find environment variable LOGDNA_KEY",
        )
    }

    custom_hostname := true
    if hostname == "" {
        host, err := os.Hostname()
        if err != nil {
            log.Fatal(err.Error())
        }
        hostname = host
        custom_hostname = false
    }

    if included == "" {
        included = []
    } else {
        included = strings.Split(included, ",")
    }

    if excluded == "" {
        excluded = []
    } else {
        excluded = strings.Split(excluded, ",")
    }

    return adapter.New(
        endpoint,
        token,
        tags,
        hostname,
        custom_hostname,
        included,
        excluded,
    ), nil
}
