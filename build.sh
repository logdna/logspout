#!/bin/sh
set -e
apk add --no-cache --update go build-base git mercurial ca-certificates
mkdir -p /go/src/github.com/gliderlabs
cp -r /src /go/src/github.com/gliderlabs/logspout
cd /go/src/github.com/gliderlabs/logspout
export GOPATH=/go
go get
go build -ldflags "-X main.Version=$1" -o /bin/logspout
