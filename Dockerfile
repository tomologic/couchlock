FROM debian:jessie

ADD . /usr/src/app
WORKDIR /usr/src/app

RUN set -x; buildDeps="curl golang make git"; \
    # Install runtime and build dependencies
    apt-get update && \
    apt-get install -y ca-certificates $buildDeps && \
    # Install travis-ci gimme
    curl -sL -o /usr/local/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme && \
    chmod +x /usr/local/bin/gimme && \
    # Build binary
    make build_linux && \
    ln -s $PWD/ARTIFACTS/couchlock-*-linux-amd64 /usr/local/bin/couchlock && \
    # Cleanup
    apt-get purge -y $buildDeps && \
    apt-get autoremove -y --purge && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENTRYPOINT ["/usr/local/bin/couchlock"]
