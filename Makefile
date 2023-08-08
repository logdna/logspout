BUILD_ARCH  := $(shell uname -m)
BUILD_OS    := $(shell uname -s | tr '[:upper:]' '[:lower:]')
TAG          = $(shell VER=$$(git tag | tail -1); echo $${VER:1})
REPO         = ${USERNAME}
NAME         = logspout
IMAGE        = $(REPO)/$(NAME)
IMAGE_AMD64  = $(IMAGE):$(TAG)-amd64
IMAGE_ARM64  = $(IMAGE):$(TAG)-arm64

ifeq ($(BUILD_ARCH), x86_64)
BUILD_ARCH = amd64
endif

ifndef GIT_BRANCH
GIT_BRANCH := $(shell git branch --show-current)
endif

SVU_ARCH=$(BUILD_ARCH)
ifeq ($(BUILD_OS), darwin)
SVU_ARCH=all
endif

NEXT := next --force-patch-increment
ifeq ($(GIT_BRANCH), $(DEFAULT_BRANCH))
SVU_OPTS := $(NEXT)
else
SVU_OPTS := current
endif

# Enable 'docker manifest' commands (still experimental in Docker v20.10.5)
export DOCKER_CLI_EXPERIMENTAL = enabled

.PHONY: lint
.SILENT: lint
lint:
	n=$$(gofmt -l logdna | wc -l); \
	if [ $$n -gt 0 ]; then \
	    echo 'run "make lint-fix" to fix the following files with formatting errors:'; \
	    gofmt -l .; \
		echo ; \
		echo 'or run "make lint-diff" to see the changes it wants to make'; \
	    exit 1; \
	fi

.PHONY: build
build:
	docker build --pull -t $(IMAGE_AMD64) \
		--build-arg ARCH=amd64 \
		--build-arg OS=linux \
		-f Dockerfile .
	docker build --pull -t $(IMAGE_ARM64) \
		--build-arg ARCH=arm64 \
		--build-arg OS=linux \
		-f Dockerfile .
	docker save -o image.tar $(IMAGE_AMD64) $(IMAGE_ARM64)

.PHONY: publish
publish:
	docker load -i ./image.tar
	docker push $(IMAGE_AMD64)
	docker push $(IMAGE_ARM64)
	docker manifest create $(IMAGE):$(TAG) $(IMAGE_AMD64) $(IMAGE_ARM64)
	docker manifest annotate $(IMAGE):$(TAG) $(IMAGE_ARM64) --arch arm64 --os linux
	docker manifest push --purge $(IMAGE):$(TAG)
	docker manifest create $(IMAGE):latest $(IMAGE_AMD64) $(IMAGE_ARM64)
	docker manifest annotate $(IMAGE):latest $(IMAGE_ARM64) --arch arm64 --os linux
	docker manifest push --purge $(IMAGE):latest

version: | svu ## tag a new version
	git fetch --tags
	./svu $(SVU_OPTS) | tee version
	if [[ "$(GIT_BRANCH)" == "$(DEFAULT_BRANCH)" ]]; then \
		git tag $$(cat version); \
		git push --tags; \
	fi

svu:
	latest_real=$$(curl -sfL -o /dev/null -w %{url_effective} https://github.com/caarlos0/svu/releases/latest | rev | cut -f 1 -d / | rev | tr -d v); \
	curl -sL https://github.com/caarlos0/svu/releases/download/v$${latest_real}/svu_$${latest_real}_$(BUILD_OS)_$(SVU_ARCH).tar.gz | \
	tar -zx svu