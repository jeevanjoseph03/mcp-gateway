.PHONY: build run test clean

build:
	go build -o bin/mcp-gateway cmd/gateway/main.go

run: build
	./bin/mcp-gateway -config configs/config.yaml

test:
	go test -v ./...

clean:
	rm -rf bin/

deps:
	go mod download
	go mod tidy

# For development - runs with hot reload (requires air)
dev:
	air -c .air.toml