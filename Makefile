-include .config.mk

all: build

ALL_FILES := $(shell find . -type f -name '*.go')
GO_FILES := $(sort $(filter-out %_test.go, $(ALL_FILES)))
HAVE_TESTS := $(sort $(filter %_test.go, $(ALL_FILES)))
WANT_TESTS := $(sort $(patsubst %.go,%_test.go, $(GO_FILES)))
# lack of spaces after commas in WANT_TESTS is important
# for formatting error output in `make test`

APP_NAME := logspout

BUILD_ARCH := $(shell uname -m)
BUILD_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

GORELEASER_ARCH=$(BUILD_ARCH)
ifeq ($(BUILD_ARCH), x86_64)
BUILD_ARCH = amd64
GORELEASER_ARCH = x86_64
endif

SVU_ARCH=$(BUILD_ARCH)
ifeq ($(BUILD_OS), darwin)
SVU_ARCH=all
endif

ifndef GIT_BRANCH
GIT_BRANCH := $(shell git branch --show-current)
endif

ifdef JENKINS_URL
build: build-all
test: test-ci
else
build: build-one
test: test-local
endif

NEXT := next --force-patch-increment
ifeq ($(GIT_BRANCH), main)
PUBLISH_OPTS := --clean
SVU_OPTS := $(NEXT)
else
PUBLISH_OPTS := --clean --snapshot --skip-publish
SVU_OPTS := current
endif

BUILD_CMD ?= ./goreleaser build --clean --snapshot
BUILD_ONE_OPTS ?= --single-target --output .
COVERAGE ?= $(REPORT_DIR)/coverage
ifndef GOPATH
	GOPATH := $(shell pwd)/.go
endif
GOBIN ?= $(GOPATH)/bin
GOCOV ?= $(GOBIN)/gocov
GOCOVXML ?= $(GOBIN)/gocov-xml
GOCOVHTML ?= $(GOBIN)/gocov-html
REPORT_DIR ?= reports

-include logspout.mk

.PHONY: debug-%
debug-%:              ## Debug a variable by calling `make debug-VARIABLE`
	@echo $(*) = $($(*))

.PHONY: help
.SILENT: help
help:                 ## Show this help, includes list of all actions.
	@awk 'BEGIN {FS = ":.*?## "}; /^.+: .*?## / && !/awk/ {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ${MAKEFILE_LIST} | sort

build: ## in CI: build for all platforms. locally: build for your platform

build-all: | goreleaser ## build linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 (macOS is darwin)
	$(BUILD_CMD)

build-one: lint-fix | goreleaser ## build one binary for the platform you're running on
	GOARCH=$(BUILD_ARCH) GOOS=$(BUILD_OS) $(BUILD_CMD) $(BUILD_ONE_OPTS)

clean: clean-coverage clean-test ## clean dist/ && binary
	-rm -rf dist/
	-find . -maxdepth 1 -name '$(APP_NAME)' -delete
	-find . -maxdepth 1 -name version -delete
	go clean -cache

clean-all: clean ## clean, clean test caches, remove tooling (svu, etc)
	-find . -maxdepth 1 -name 'goreleaser' -delete
	-find . -maxdepth 1 -name 'svu' -delete

clean-coverage: ## clean coverage files
	-rm -rf $(REPORT_DIR)

clean-test: ## reset test cache
	go clean -testcache

.PHONY: coverage-html
coverage-html: $(COVERAGE).html ## generate html coverage report

.PHONY: coverage-json
coverage-json: $(COVERAGE).json ## generate json coverage report

.PHONY: coverage-threshold
.SILENT: coverage-threshold
coverage-threshold: $(COVERAGE).txt ## if set, test coverage is COVERAGE_THRESHOLD or higher
	COVERAGE_CURRENT=$(shell go tool cover -func $(COVERAGE).txt | grep total | awk '{print $$3}' | sed 's/%//g'); \
	INSUFFICIENT_COVERAGE=$$(echo "$$COVERAGE_CURRENT < $(COVERAGE_THRESHOLD)" | bc); \
	if [ $$INSUFFICIENT_COVERAGE -eq 1 ] ; then \
		echo "ERROR insufficent test coverage!"; \
		echo "this repo requires $(COVERAGE_THRESHOLD)% coverage"; \
		echo "it currently has $$COVERAGE_CURRENT%"; \
		echo "run 'make show-coverage' to see a coverage report locally"; \
		exit 1; \
	fi

.PHONY: coverage-txt
coverage-txt: $(COVERAGE).txt ## generate plaintext coverage report

.PHONY: coverage-xml
coverage-xml: $(COVERAGE).xml ## generate xml coverage report

.PHONY: show-coverage
show-coverage: clean $(COVERAGE).html ## generate and open html coverage report
	open $(COVERAGE).html

$(COVERAGE).html: $(COVERAGE).json | $(GOCOVHTML)
	cat $(COVERAGE).json | $(GOCOVHTML) > $(@)

$(COVERAGE).json: $(COVERAGE).txt | $(GOCOV)
	$(GOCOV) convert $(COVERAGE).txt > $(@)

$(COVERAGE).txt: | $(REPORT_DIR)
	go test -cover -coverprofile=$(@) ./...

$(COVERAGE).xml: $(COVERAGE).json | $(GOCOVXML)
	cat $(COVERAGE).json | $(GOCOVXML) > $(@)

$(GOPATH) $(GOBIN) $(REPORT_DIR):
	mkdir -p $@

.PHONY: lint
.SILENT: lint
lint: lint-goreleaser ## exit 1 if gofmt would change the code
	n=$$(gofmt -l . | wc -l); \
	if [ $$n -gt 0 ]; then \
	    echo 'run "make lint-fix" to fix the following files with formatting errors:'; \
	    gofmt -l .; \
		echo ; \
		echo 'or run "make lint-diff" to see the changes it wants to make'; \
	    exit 1; \
	fi

.PHONY: lint-diff
lint-diff: ## use gofmt to show the diff between what you have and what it wants
	gofmt -d .

.PHONY: lint-fix
lint-fix: ## use gofmt to reformat the code
	gofmt -w .

.PHONY: lint-goreleaser
lint-goreleaser: | goreleaser
	./goreleaser check .goreleaser.yaml

publish: version | goreleaser ## build for all platforms, generate changelog, cut a github release, add binaries as release assets
	./goreleaser release $(PUBLISH_OPTS)

.SILENT: test
test: test-files-exist ## run test-ci or test-local as appropriate

.PHONY: test-files-exist
.SILENT: test-files-exist
test-files-exist: ## ensure *.go files in repo root, cmd, pkg have _test.go files
ifneq ($(HAVE_TESTS), $(WANT_TESTS))
	echo
	echo "test files do not match expected *_test.go pattern!"
	echo "based on these go files:  $(GO_FILES)"
	echo
	echo "we want these test files: $(WANT_TESTS)"
	echo
	echo "we have these test files: $(HAVE_TESTS)"
	echo
	exit 1
endif

test-ci: coverage-html coverage-xml ## for CI: run tests and generate coverage reports

ifdef COVERAGE_THRESHOLD
test-ci: coverage-threshold
endif

test-local: lint-fix ## run go tests locally
	go test ./...

version: | svu ## tag a new version
	git fetch --tags
	./svu $(SVU_OPTS) | tee version
	if [[ "$(GIT_BRANCH)" == "main" ]]; then \
		git tag $$(cat version); \
		git push --tags; \
	fi

version-current: | svu
	git fetch --tags
	./svu current

version-next: | svu
	git fetch --tags
	./svu $(NEXT)

.PHONY: gocov
gocov: | $(GOCOV)

$(GOCOV): | $(GOBIN)
	GOPATH=$(GOPATH) go install github.com/axw/gocov/gocov@latest

.PHONY: gocov-html
gocov-html: | $(GOCOVHTML)

$(GOCOVHTML): | $(GOBIN)
	GOPATH=$(GOPATH) go install github.com/matm/gocov-html/cmd/gocov-html@latest

.PHONY: gocov-xml
gocov-xml: | $(GOCOVXML)

$(GOCOVXML): | $(GOBIN)
	GOPATH=$(GOPATH) go install github.com/AlekSi/gocov-xml@latest

goreleaser:
	latest_real=$$(curl -sfL -o /dev/null -w %{url_effective} https://github.com/goreleaser/goreleaser/releases/latest | rev | cut -f 1 -d / | rev | tr -d v); \
	curl -sL https://github.com/goreleaser/goreleaser/releases/download/v$${latest_real}/goreleaser_$(BUILD_OS)_$(GORELEASER_ARCH).tar.gz | \
	tar -zx goreleaser

svu:
	latest_real=$$(curl -sfL -o /dev/null -w %{url_effective} https://github.com/caarlos0/svu/releases/latest | rev | cut -f 1 -d / | rev | tr -d v); \
	curl -sL https://github.com/caarlos0/svu/releases/download/v$${latest_real}/svu_$${latest_real}_$(BUILD_OS)_$(SVU_ARCH).tar.gz | \
	tar -zx svu
