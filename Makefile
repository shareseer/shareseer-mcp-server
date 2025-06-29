# ShareSeer MCP Server Makefile

BINARY=shareseer-mcp
GO=go
COVERTEMP=c.out
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

.PHONY: clean deps build test all run coverage bench docker release

all: deps build test

deps:
	$(GO) mod tidy
	$(GO) mod download

build:
	$(GO) build -v $(LDFLAGS) -o $(BINARY) ./cmd/server

test: build
	$(GO) test -v ./...

coverage:
	$(GO) test -cover -coverprofile=$(COVERTEMP) ./...
	$(GO) tool cover -html=$(COVERTEMP)
	@rm -f $(COVERTEMP)

clean:
	$(GO) clean
	@rm -f $(BINARY) $(BINARY)-*

run: build
	./$(BINARY)

dev: build
	./$(BINARY) &
	@echo "ShareSeer MCP server running on http://localhost:8081"
	@echo "Visit http://localhost:8081/mcp/info for available tools"

# Release builds for multiple platforms
release: clean
	@echo "Building release binaries..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BINARY)-darwin-arm64 ./cmd/server
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY)-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BINARY)-linux-arm64 ./cmd/server
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY)-windows-amd64.exe ./cmd/server
	@echo "Release binaries built successfully"

# Docker commands
docker-build:
	docker build --build-arg VERSION=$(VERSION) -t shareseer/mcp-server:$(VERSION) .
	docker tag shareseer/mcp-server:$(VERSION) shareseer/mcp-server:latest

docker-run:
	docker run -p 8081:8081 \
		-e REDIS_ADDR=localhost:6379 \
		shareseer/mcp-server:latest

docker-push:
	docker push shareseer/mcp-server:$(VERSION)
	docker push shareseer/mcp-server:latest

# Production deployment
deploy-prod:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY)-linux ./cmd/server
	@echo "Upload $(BINARY)-linux to your server"

# Install locally
install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp $(BINARY) /usr/local/bin/
	@echo "Installation complete"

# Development helpers
fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run

bench: build
	$(GO) test -bench=. ./...

# Generate checksums for release
checksums:
	@echo "Generating checksums..."
	sha256sum $(BINARY)-* > checksums.txt
	@echo "Checksums saved to checksums.txt"