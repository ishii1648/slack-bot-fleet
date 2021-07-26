#!/usr/bin/env bash

set -e

COVERAGE=$1

echo "" > coverage.txt

for d in $(go list ./...); do
    if [[ $COVERAGE = "-coverage" ]]; then
        go test -v -race -coverprofile=profile.out -covermode=atomic $d
        if [ -f profile.out ]; then
            cat profile.out >> coverage.txt
            rm profile.out
        fi
    else
        go test -v -race $d
    fi
done