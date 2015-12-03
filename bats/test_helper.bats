#!/usr/bin/env bats

load "test_helper"

setup () {
    setup_couchdb
}

teardown () {
    cleanup_couchdb
}

@test "create couchdb and test database" {
    run curl -X PUT $COUCHDB/test/
    [ "$status" -eq 0 ]

    run curl -X GET $COUCHDB/test/
    [ "$status" -eq 0 ]
}

@test "create couchdb and test database twice" {
    run curl -X PUT $COUCHDB/test/
    [ "$status" -eq 0 ]

    run curl -X GET $COUCHDB
    [ "$status" -eq 0 ]
}
