#!/usr/bin/env bats

load "test_helper"

setup () {
    setup_couchdb

    run curl -X PUT $COUCHDB/test/
    [ "$status" -eq 0 ]
}

teardown () {
    cleanup_couchdb
}

@test "simple lock" {
    run couchlock --couchdb $COUCHDB/test/ \
        --lock test-lock \
        --name owner1 \
        lock
    [ "$status" -eq 0 ]

    run couchlock --couchdb $COUCHDB/test/ \
        --lock test-lock \
        --name owner1 \
        unlock
    [ "$status" -eq 0 ]
}
