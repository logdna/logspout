package logdna

import (
	"fmt"
	"io/ioutil"
	"logdna/adapter"
	"os"
	"strings"
	"testing"

	"github.com/gliderlabs/logspout/router"
	"github.com/stretchr/testify/assert"
)

func TestHostname(t *testing.T) {
	originalKey := os.Getenv("LOGDNA_KEY")
	assert := assert.New(t)

	t.Run("NewLogDNAAdapter func requires LOGDNA_KEY", func(t *testing.T) {
		os.Setenv("LOGDNA_KEY", "")
		result, err := NewLogDNAAdapter(&router.Route{})
		assert.Nil(result)
		assert.Error(err)
		assert.Equal(err.Error(), "Cannot Find Environment Variable \"LOGDNA_KEY\"")
	})

	os.Setenv("LOGDNA_KEY", "dummy_key")

	t.Run("NewLogDNAAdapter sets LogDNAKey config", func(t *testing.T) {
		result, err := NewLogDNAAdapter(&router.Route{})
		assert.NotNil(result)
		assert.NoError(err)
		adapter := result.(*adapter.Adapter)
		assert.NotNil(adapter)
		assert.Equal(os.Getenv("LOGDNA_KEY"), adapter.Config.LogDNAKey)
	})

	t.Run("NewLogDNAAdapter sets tags config", func(t *testing.T) {
		os.Setenv("TAGS", "abc,def")
		result, err := NewLogDNAAdapter(&router.Route{})
		assert.NotNil(result)
		assert.NoError(err)
		adapter := result.(*adapter.Adapter)
		assert.NotNil(adapter)
		assert.Equal(os.Getenv("TAGS"), adapter.Config.Tags)
	})

	if _, e := os.Stat("/etc/host_hostname"); !os.IsNotExist(e) {
		t.Run("NewLogDNAAdapter func uses etc hosthost name when defined", func(t *testing.T) {
			originalEtcHostName, _ := ioutil.ReadFile("/etc/host_hostname")
			fmt.Println("--Original", originalEtcHostName)
			expected := strings.TrimRight(string(originalEtcHostName), "\r\n")
			fmt.Println("--Expected", expected, "-")

			route := router.Route{}
			result, err := NewLogDNAAdapter(&route)
			assert.NoError(err)
			assert.NotNil(result)
			adapter := result.(*adapter.Adapter)
			assert.NotNil(adapter)
			fmt.Println("--Obtained", adapter.Config.Hostname, "-")
			assert.Equal(expected, adapter.Config.Hostname)
		})
	} else {
		t.Run("NewLogDNAAdapter func uses host name when defined", func(t *testing.T) {
			if os.Getenv("HOSTNAME") == "" {
				os.Setenv("HOSTNAME", "my-host-name")
			}

			route := router.Route{}
			result, err := NewLogDNAAdapter(&route)
			assert.NoError(err)
			assert.NotNil(result)
			adapter := result.(*adapter.Adapter)
			assert.NotNil(adapter)
			assert.Equal(os.Getenv("HOSTNAME"), adapter.Config.Hostname)
		})
	}

	// Restore values
	os.Setenv("LOGDNA_KEY", originalKey)
}
