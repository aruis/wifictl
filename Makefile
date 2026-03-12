BINARY := wifictl
DIST_DIR := dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -X github.com/aruis/wifictl/internal/buildinfo.Version=$(VERSION) -X github.com/aruis/wifictl/internal/buildinfo.Commit=$(COMMIT) -X github.com/aruis/wifictl/internal/buildinfo.Date=$(BUILD_TIME)

.PHONY: build test clean

build:
	mkdir -p $(DIST_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY) ./

test:
	go test ./...

clean:
	rm -rf $(DIST_DIR)
