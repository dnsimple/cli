BINARY_NAME := dnsimple
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build install test lint clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/dnsimple

install:
	go install $(LDFLAGS) ./cmd/dnsimple

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
