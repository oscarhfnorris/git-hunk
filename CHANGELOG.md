# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project scaffold with Go module structure.
- `internal/hunk` — unified-diff hunk extraction and SHA-256 content hashing.
- `internal/storage` — read/write layer for `refs/meta/hunks` Git ref tree.
- `internal/llm` — provider-agnostic LLM client interface with OpenAI implementation.
- `internal/config` — TOML config file loading from `~/.config/git-hunk/config.toml`.
- `cmd/git-hunk/main.go` — CLI entry point with subcommands:
  `generate`, `show`, `push`, `fetch`, `verify`, `install-hooks`.
- `schema/v1.json` — JSON Schema for hunk metadata blobs.
- `spec/STORAGE.md` — storage protocol specification.
- `hooks/prepare-commit-msg` and `hooks/post-rewrite` — Git hook scripts.
- `buildScripts/` — PowerShell 7 build, test, lint, check-deps, pester scripts.
- `buildScripts/*.Tests.ps1` — Pester 5 test suites for each build script.
- `.PSScriptAnalyzerSettings.psd1` — PSScriptAnalyzer settings.
- `Makefile` — Linux/macOS make targets mirroring the PowerShell scripts.
- `.vscode/` — tasks, launch configs, and extension recommendations.
- `.github/workflows/` — CI, PR rules, and PowerShell test workflows.
- `.editorconfig`, `.gitignore`, `.env.example` — repo hygiene files.
- `testdata/` — sample unified diff fixtures for unit tests.
