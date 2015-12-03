#!/usr/bin/env bats

@test "version subcommand" {
    run couchlock version
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^[0-9]*\.[0-9]*\.[0-9]*.*$ ]]
}
