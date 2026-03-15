# Makefile for git-hunk
# Mirrors the PowerShell buildScripts/ for Linux/macOS developers.

BINARY      := git-hunk
BUILD_DIR   := dist
COVERAGE    := coverage.out
MODULE      := github.com/oscarhfnorris/git-hunk

.PHONY: all build test test-cover lint lint-fix install clean check-deps pester

all: build

## build: compile all packages
build:
	go build ./...

## release: produce an optimised binary in ./dist/
release:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) ./cmd/git-hunk/...

## test: run all Go tests
test:
	go test ./...

## test-cover: run tests and write coverage.out
test-cover:
	go test -coverprofile=$(COVERAGE) -covermode=atomic ./...
	go tool cover -func=$(COVERAGE)

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## lint-fix: run golangci-lint with --fix
lint-fix:
	golangci-lint run --fix ./...

## install: install the binary to GOPATH/bin
install:
	go install ./cmd/git-hunk/...

## clean: remove build artefacts
clean:
	rm -rf $(BUILD_DIR) $(COVERAGE) coverage-ps.xml

## check-deps: check for outdated modules and vulnerabilities
check-deps:
	go list -u -m all
	govulncheck ./...

## pester: run Pester tests (requires PowerShell 7)
pester:
	pwsh -NoLogo -NonInteractive -File buildScripts/run-pester.ps1
