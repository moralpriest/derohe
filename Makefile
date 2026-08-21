GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

.PHONY: all build build-all test test-simulator test-integration fmt lint clean

all: build

build:
	@mkdir -p bin
	@echo "Building derod..."
	go build -mod=vendor -trimpath -ldflags=-buildid= -o bin/derod ./cmd/derod
	@echo "Building dero-wallet-cli..."
	go build -mod=vendor -trimpath -ldflags=-buildid= -o bin/dero-wallet-cli ./cmd/dero-wallet-cli
	@echo "Building dero-miner..."
	go build -mod=vendor -trimpath -ldflags=-buildid= -o bin/dero-miner ./cmd/dero-miner
	@echo "Building explorer..."
	go build -mod=vendor -trimpath -ldflags=-buildid= -o bin/explorer ./cmd/explorer
	@echo "Building simulator..."
	go build -mod=vendor -trimpath -ldflags=-buildid= -o bin/simulator ./cmd/simulator
	@echo "All binaries built in bin/"

build-all:
	@echo "Building all binaries for all platforms..."
	./build_all.sh

test:
	@echo "Running unit tests (excluding simulator)..."
	go test -mod=vendor -timeout 30m $(shell go list -mod=vendor ./... | grep -v '/cmd/simulator$$')

test-simulator:
	@echo "Running simulator chain tests..."
	go test -mod=vendor -timeout 30m -v ./cmd/simulator/...

test-integration:
	@echo "Running simulator-based integration suite..."
	./run_integration_test.sh

fmt:
	gofmt -w -s $(shell find . -name "*.go" -not -path "./vendor/*")

lint:
	golangci-lint run --timeout=10m --exclude-dirs=^vendor/

clean:
	rm -rf bin/
	rm -f derod dero-wallet-cli dero-miner explorer simulator
	rm -f coverage.out coverage.html