GOFILES := $(shell find . -name "*.go")

.PHONY: check
check: fmt lint test

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	goimports -w -local github.com/hara/roomheatmap $(GOFILES)

.PHONY: build
build:
	goreleaser build --single-target --snapshot --rm-dist
