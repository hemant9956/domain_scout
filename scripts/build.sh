#!/usr/bin/env sh
set -eu
go test ./...
go build -o bin/domain-scout ./cmd/api
