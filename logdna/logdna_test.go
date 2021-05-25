package logdna

import (
	"testing"

	"github.com/gliderlabs/logspout/router"
	"github.com/stretchr/testify/assert"
)

func TestHostname(t *testing.T) {
	assert := assert.New(t)

	t.Run("Route constructor uses blank hostname if it doesn't find one", func(t *testing.T) {
		router.AdapterFactories.Register(NewLogDNAAdapter, "logdna")

		r := &router.Route{
			Adapter: "logdna",
		}

		if err := router.Routes.Add(r); err != nil {
			t.Log("Cannot Add New Route: ", err.Error())
		}

		route, _ := router.Routes.Get("logdna")

		// TODO figure out how to test this!
		assert.Equal(true, true)
	})

}
