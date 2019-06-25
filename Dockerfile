##########################
# Building image
##########################
FROM --platform=$BUILDPLATFORM golang:1.12-alpine3.10 AS build
ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"

# v3.2.6
ENV LOGSPOUT_VERSION=591787f3d3202cbe029a9cff6c14e3a178e1f78c

# Update and get bash and git
RUN apk add --update bash git
# Checkout logspout upstream
WORKDIR /go/src/github.com/gliderlabs/logspout
RUN git clone https://github.com/gliderlabs/logspout.git .
RUN git checkout $LOGSPOUT_VERSION
# Install glide and run it
RUN go get github.com/Masterminds/glide
RUN $GOPATH/bin/glide install

# Add the logdna stuff
COPY modules.go .
COPY logdna ./vendor/github.com/logdna/logspout/logdna

# Build it
RUN arch=${TARGETPLATFORM#*/} && arch=${arch%/*} \
    && env GOOS=linux GOARCH=$arch go build -ldflags "-X main.Version=$(cat VERSION)-custom" -o /bin/logspout

##########################
# Shipping image
##########################
FROM alpine:3.10

# Backward compatible...
COPY --from=build /bin/logspout /bin/logspout

ENTRYPOINT ["/bin/logspout"]
VOLUME /mnt/routes
EXPOSE 80
