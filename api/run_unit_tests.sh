#!/bin/bash
# start these tests from the root of the api module

# https://stackoverflow.com/questions/43580131/exec-gcc-executable-file-not-found-in-path-when-trying-go-build
export CGO_ENABLED=0

echo "Running unit tests" 
go test ./...
