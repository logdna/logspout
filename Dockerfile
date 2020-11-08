FROM gliderlabs/logspout:latest AS builder


FROM alpine:3.12
COPY --from=builder /bin/logspout /bin/logspout
ENV BUILD_VERSION=1.2.0
VOLUME /mnt/routes
RUN ln -fs /tmp/docker.sock /var/run/docker.sock
ENTRYPOINT ["/bin/logspout"]
