#!/usr/bin/env bash

# Exclude packages that require integration tests
go test $(go list ./internal/... | grep -v /payment-service | grep -v /payment-demon) -coverprofile cover.out
if (( $? != 0 )); then
    echo "Tests failed, aborting"
    exit 1
fi

coverage=$(go tool cover -func cover.out | awk 'END { sub(/%/,"",$3); print $3 }')
if (( $(echo "$coverage" | awk '{print $1 < 30}') )); then
    echo "Test coverage is below 30%, aborting"
    exit 1
else
    echo "Test coverage is above 30%, test continuing to build stage"
fi
