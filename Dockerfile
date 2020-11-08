FROM gliderlabs/logspout:latest AS builder
RUN apk add --no-cache --update go build-base git mercurial ca-certificates
RUN mkdir -p /go/src/github.com/gliderlabs && \
    cp -r /src /go/src/github.com/gliderlabs/logspout
WORKDIR /go/src/github.com/gliderlabs/logspout
ENV GOPATH=/go
RUN go get
RUN go build -ldflags "-X main.Version=$(cat VERSION)-logdna" -o /bin/logspout


FROM alpine:3.12
COPY --from=builder /bin/logspout /bin/logspout
ENV BUILD_VERSION=1.2.0
VOLUME /mnt/routes
RUN ln -fs /tmp/docker.sock /var/run/docker.sock
ENTRYPOINT ["/bin/logspout"]
