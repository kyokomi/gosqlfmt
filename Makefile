GOLANGCI_LINT_VERSION := v2.9.0

.PHONY: lint test build

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

test:
	go test -race ./...

build:
	go build ./...
