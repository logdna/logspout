TAG          = $(if $(CIRCLE_TAG),$(CIRCLE_TAG),PRTEST)
REPO         = ${USERNAME}
NAME         = logspout
IMAGE        = $(REPO)/$(NAME)
IMAGE_AMD64  = $(IMAGE):$(TAG)-amd64
IMAGE_ARM64  = $(IMAGE):$(TAG)-arm64

# Enable 'docker manifest' commands (still experimental in Docker v20.10.5)
export DOCKER_CLI_EXPERIMENTAL = enabled

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

publish:
	docker login --username ${USERNAME} --password ${PASSWORD}
	docker load -i ./image.tar
	docker push $(IMAGE_AMD64)
	docker push $(IMAGE_ARM64)
	docker manifest create $(IMAGE):$(TAG) $(IMAGE_AMD64) $(IMAGE_ARM64)
	docker manifest annotate $(IMAGE):$(TAG) $(IMAGE_ARM64) --arch arm64 --os linux
	docker manifest push --purge $(IMAGE):$(TAG)
	docker manifest create $(IMAGE):latest $(IMAGE_AMD64) $(IMAGE_ARM64)
	docker manifest annotate $(IMAGE):latest $(IMAGE_ARM64) --arch arm64 --os linux
	docker manifest push --purge $(IMAGE):latest
