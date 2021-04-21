FROM gliderlabs/logspout:latest AS builder
RUN apk add --no-cache --update go build-base git mercurial ca-certificates
RUN mkdir -p /go/src/github.com/gliderlabs && \
    cp -r /src /go/src/github.com/gliderlabs/logspout
WORKDIR /go/src/github.com/gliderlabs/logspout
ARG ARCH=amd64
ARG OS=linux
ENV GOARCH=${ARCH}
ENV GOOS=${OS}
ENV GOPATH=/go
ENV CGO_ENABLED=0
RUN go get
RUN go build -ldflags "-X main.Version=$(cat VERSION)-logdna" -o /bin/logspout


FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/logspout /bin/logspout
ENV BUILD_VERSION=1.3.1
VOLUME /mnt/routes
ENTRYPOINT ["/bin/logspout"]
