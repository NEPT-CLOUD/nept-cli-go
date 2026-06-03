BINARY_NAME=nept
VERSION?=1.0.0
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS=-ldflags "-X github.com/NEPT-CLOUD/nept-cli-go/cmd.Version=$(VERSION) \
                  -X github.com/NEPT-CLOUD/nept-cli-go/cmd.GitCommit=$(COMMIT) \
                  -X github.com/NEPT-CLOUD/nept-cli-go/cmd.BuildDate=$(BUILD_DATE)"

.PHONY: all build test schema tidy clean install help

all: build

build: ## Build the binary with version and git commit details injected
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

test: ## Run unit tests with coverage and race detection
	go test -v -race -coverprofile=coverage.out ./...

schema: ## Generate and print the CLI command schema in JSON
	go run main.go schema --format json

tidy: ## Clean up and sync go.mod/go.sum
	go mod tidy

clean: ## Remove build artifacts and coverage records
	rm -f $(BINARY_NAME) coverage.out

install: ## Install the binary globally in $GOPATH/bin
	go install $(LDFLAGS)

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
