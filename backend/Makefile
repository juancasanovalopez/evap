SHELL := /bin/zsh
GOOS ?= linux
GOARCH ?= arm64
BINARY := bootstrap
DIST := dist

.PHONY: test test-integration lint build package clean

test:
	go test ./...

test-integration:
	go test -tags=integration ./test/integration/...

lint:
	golangci-lint run ./...

build:
	mkdir -p $(DIST)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -o $(DIST)/$(BINARY) ./cmd/api

package: build
	cd $(DIST) && zip -j function.zip $(BINARY)

clean:
	rm -rf $(DIST)
