.PHONY: run test build build-agent build-linux-amd64 build-linux-arm64 build-windows-amd64 fmt vet clean

run:
	go run ./cmd/api

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/domain-scout-api ./cmd/api

build-agent:
	mkdir -p dist
	go build -trimpath -ldflags="-s -w" -o dist/domain-scout ./cmd/agent

build-linux-adm64:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/domain-scout-linux-amd64 ./cmd/agent

build-linux-arm64:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/domain-scout-linux-arm64 ./cmd/agent

build-windows-amd64:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/domain-scout-windows-amd64.exe ./cmd/agent

fmt:
	gofmt -w cmd internal pkg

vet:
	go vet ./...

clean:
	rm -rf bin dist
