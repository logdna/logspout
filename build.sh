#!/bin/sh
set -e
echo "set -e"
apk add --update go build-base git mercurial ca-certificates
echo "apk add --update go build-base git mercurial ca-certificates"
mkdir -p /go/src/github.com/gliderlabs
echo "mkdir -p /go/src/github.com/gliderlabs"
cp -r /src /go/src/github.com/gliderlabs/logspout
echo "cp -r /src /go/src/github.com/gliderlabs/logspout"
cd /go/src/github.com/gliderlabs/logspout
echo "cd /go/src/github.com/gliderlabs/logspout"
export GOPATH=/go
echo "export GOPATH=/go"
go get
echo "go get"
go build -ldflags "-X main.Version=$1" -o /bin/logspout
echo "go build -ldflags \"-X main.Version=$1\" -o /bin/logspout"
apk del go git mercurial build-base
echo "apk del go git mercurial build-base"
rm -rf /go /var/cache/apk/* /root/.glide
echo "rm -rf /go /var/cache/apk/* /root/.glide"

# backwards compatibility
ln -fs /tmp/docker.sock /var/run/docker.sock
echo "ln -fs /tmp/docker.sock /var/run/docker.sock"
