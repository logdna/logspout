package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	ARCH = "archie"
	OS = "jughead"
	SEMVER = "veronica"

	var buf bytes.Buffer
	log.SetOutput(&buf)
	main()
	log.SetOutput(os.Stderr)

	fail := false
	have := buf.String()
	want := []string{
		"hello world!",
		"i was created by repo-template-go",
		"veronica-jughead-archie",
	}

	for _, w := range want {
		if !strings.Contains(have, w) {
			t.Logf("\"%s\" missing from \"%s\"", w, have)
			fail = true
		}
	}
	if fail {
		t.Fail()
	}
}
