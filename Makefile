BINARY     := kubewatch
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")
LDFLAGS    := -ldflags "-X github.com/mdryaan/kubewatch-cli/pkg/version.Version=$(VERSION) \
                         -X github.com/mdryaan/kubewatch-cli/pkg/version.GitCommit=$(GIT_COMMIT) \
                         -X github.com/mdryaan/kubewatch-cli/pkg/version.BuildDate=$(BUILD_DATE)"

.PHONY: build clean test lint fmt vet tidy install run

build:
	go build $(LDFLAGS) -o $(BINARY) ./...

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 ./...

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64 ./...

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-windows-amd64.exe ./...

build-all: build-linux build-darwin build-windows

install:
	go install $(LDFLAGS) ./...

test:
	go test ./... -v -race -cover

test-short:
	go test ./... -short

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(BINARY) $(BINARY)-linux-amd64 $(BINARY)-darwin-amd64 $(BINARY)-windows-amd64.exe
	rm -f coverage.out coverage.html

run:
	go run ./... $(ARGS)

help:
	@echo "Available targets:"
	@echo "  build        — build for current platform"
	@echo "  build-all    — build for linux, darwin, windows"
	@echo "  install      — install to GOPATH/bin"
	@echo "  test         — run all tests"
	@echo "  coverage     — generate HTML coverage report"
	@echo "  lint         — run golangci-lint"
	@echo "  fmt          — format source code"
	@echo "  vet          — run go vet"
	@echo "  tidy         — run go mod tidy"
	@echo "  clean        — remove build artifacts"
