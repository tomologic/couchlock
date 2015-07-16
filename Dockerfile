FROM debian:wheezy

ADD *.go /build/

RUN set -x; buildDeps="golang"; \
    cd /build && \
    apt-get update && \
    apt-get install -y ca-certificates $buildDeps && \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w' couchlock.go bindata.go && \
    mv /build/couchlock /usr/bin && \
    cd / && \
    rm -rf /build && \
    apt-get purge -y $buildDeps && \
    apt-get autoremove -y --purge && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENTRYPOINT ["/usr/bin/couchlock"]
