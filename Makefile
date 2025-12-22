# Makefile for langfuse-go

GOBIN ?= $$(go env GOPATH)/bin

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## Run all tests
	go test -v ./...

.PHONY: test-race
test-race: ## Run tests with race detector
	go test -v -race ./...

.PHONY: coverage
coverage: ## Generate coverage report
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	go tool cover -html=cover.out -o=cover.html
	@echo "Coverage report generated: cover.html"

.PHONY: install-go-test-coverage
install-go-test-coverage: ## Install go-test-coverage tool
	go install github.com/vladopajic/go-test-coverage/v2@latest

.PHONY: check-coverage
check-coverage: install-go-test-coverage ## Check coverage against thresholds
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	${GOBIN}/go-test-coverage --config=./.testcoverage.yml

.PHONY: lint
lint: ## Run linter
	golangci-lint run ./...

.PHONY: build
build: ## Build the package
	go build ./...

.PHONY: clean
clean: ## Clean generated files
	rm -f cover.out cover.html

.PHONY: examples
examples: ## Run all examples (requires LANGFUSE_PUBLIC_KEY and LANGFUSE_SECRET_KEY)
	@for dir in examples/*/; do \
		echo "Running $$dir..."; \
		(cd $$dir && go run main.go) || true; \
	done

