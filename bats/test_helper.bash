#!/bin/bash

setup_couchdb() {
    COUCHDB_IMAGE="fedora/couchdb"

    # Pull couchdb image if missing
    docker history -q $COUCHDB_IMAGE > /dev/null 2>&1 || {
        docker pull $COUCHDB_IMAGE
    }

    # Start a new couchdb
    COUCHDB_ID=$(docker run -P -d $COUCHDB_IMAGE)

    # Get port for couchdb
    COUCHDB_PORT=$(docker inspect \
        -f '{{index .NetworkSettings.Ports "5984/tcp" 0 "HostPort"}}' \
        "$COUCHDB_ID")

    # Check if docker machine is in use
    if [ -n "$DOCKER_MACHINE_NAME" ]; then
        # Get host for registry
        DOCKER_MACHINE_IP=$(docker-machine ip "$DOCKER_MACHINE_NAME")

        COUCHDB="http://$DOCKER_MACHINE_IP:$COUCHDB_PORT/"
    else
        # Assume docker on 127.0.0.1
        COUCHDB="http://127.0.0.1:$COUCHDB_PORT/"
    fi

    echo "$COUCHDB"

    # Wait until couchdb ready
    wget --output-document=- \
        --retry-connrefused \
        --timeout=10 \
        "$COUCHDB" > /dev/null
}

cleanup_couchdb() {
    docker rm -f "$COUCHDB_ID"
}
