FROM dubodubonduponey/gliderlabs-logspout:v1 AS source

FROM --platform=$BUILDPLATFORM golang:1.12-alpine3.10 AS build
ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"

RUN apk add --update bash git

RUN go get github.com/Masterminds/glide

COPY --from=source /src /go/src/github.com/gliderlabs/logspout

WORKDIR /go/src/github.com/gliderlabs/logspout
COPY modules.go .
RUN go get
RUN $GOPATH/bin/glide install

RUN arch=${TARGETPLATFORM#*/} && arch=${arch%/*} \
    && env GOOS=linux GOARCH=$arch go build -ldflags "-X main.Version=$(cat VERSION)-custom" -o /bin/logspout

# Shipping image
FROM alpine:3.10
COPY --from=build /bin/logspout /bin/logspout

ENTRYPOINT ["/bin/logspout"]
VOLUME /mnt/routes
EXPOSE 80
