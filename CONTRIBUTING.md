# Contributing to git-hunk

Thank you for taking the time to contribute! This document explains how to
set up a development environment, run tests, and submit changes.

---

## Prerequisites

| Tool | Minimum version | Notes |
|------|-----------------|-------|
| Go | 1.22 | `go version` |
| PowerShell | 7.0 | `pwsh --version` (optional — for build scripts and Pester) |
| golangci-lint | latest | installed automatically by `buildScripts/lint.ps1` |
| Pester | 5.x | installed automatically by `buildScripts/run-pester.ps1` |

---

## Dev setup

```bash
# Clone the repo
git clone https://github.com/oscarhfnorris/git-hunk.git
cd git-hunk

# Download dependencies
go mod download

# Build
go build ./...

# Run tests
go test ./...
```

---

## Running Go tests

```bash
# All tests
go test ./...

# With coverage
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out   # open in browser

# A specific package
go test ./internal/hunk/...
```

Or use the PowerShell helper:

```pwsh
./buildScripts/test.ps1 -Coverage
```

---

## Running Pester tests

Pester 5 is required and is installed automatically if missing.

```pwsh
./buildScripts/run-pester.ps1
./buildScripts/run-pester.ps1 -CI -Coverage
```

---

## Linting

### Go

```pwsh
./buildScripts/lint.ps1        # check
./buildScripts/lint.ps1 -Fix   # auto-fix
```

Or directly:

```bash
golangci-lint run ./...
```

### PowerShell

```pwsh
./buildScripts/lint-powershell.ps1
```

---

## PR process

1. Branch names must follow: `feat/my-feature`, `fix/the-bug`, `docs/update`, etc.
2. PR titles must follow [Conventional Commits](https://www.conventionalcommits.org/).
3. All Go tests must pass (`go test ./...`).
4. Lint must pass (`golangci-lint run ./...`).
5. Every new internal package must have a `_test.go` file.
6. Update `CHANGELOG.md` with a summary under `[Unreleased]`.

---

## Code structure

```
cmd/git-hunk/      Entry point and subcommand wiring
internal/config/   Config loading (TOML)
internal/hunk/     Hunk extraction from diffs + content hashing
internal/storage/  refs/meta/hunks read/write layer
internal/llm/      LLM client interface + OpenAI implementation
schema/            JSON Schema for metadata blobs
spec/              Storage protocol spec
hooks/             Git hook scripts
buildScripts/      PowerShell build tooling
testdata/          Sample unified-diff fixtures
```
