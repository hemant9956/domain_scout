.PHONY: run test build fmt vet

run:
	go run ./cmd/api

test:
	go test ./...

build:
	go build -o bin/domain-scout ./cmd/api

fmt:
	gofmt -w cmd internal pkg

vet:
	go vet ./...
