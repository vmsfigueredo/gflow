BIN      := bin/gflow
MAIN     := ./cmd/gflow
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -s -w -X main.version=$(VERSION)

.PHONY: build test lint clean release-snapshot

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) $(MAIN)

test:
	go test -race ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

release-snapshot:
	goreleaser release --snapshot --clean
