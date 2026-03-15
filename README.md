# git-hunk

Persistent, per-hunk semantic metadata for Git commits — content-addressed,
LLM-generated, and stored as a native Git ref so it pushes and fetches like
normal code.

[![CI](https://github.com/oscarhfnorris/git-hunk/actions/workflows/ci.yml/badge.svg)](https://github.com/oscarhfnorris/git-hunk/actions/workflows/ci.yml)

---

## What it does

git-hunk attaches a **named, searchable summary** to every changed block
(hunk) in every commit.  The metadata is:

* **content-addressed** — stable across rebases, amends, and cherry-picks
* **stored in the repo** — under `refs/meta/hunks` so it travels with the code
* **LLM-generated** — plain-text summaries from OpenAI (or any compatible API)
* **schema-versioned** — JSON blobs validated against `schema/v1.json`

---

## Why

`git log --oneline` tells you *what* changed; git-hunk tells you *why each
piece changed*.  Useful for code review, onboarding, and automated tooling.

---

## Install

```bash
go install github.com/oscarhfnorris/git-hunk/cmd/git-hunk@latest
```

Or build from source:

```bash
git clone https://github.com/oscarhfnorris/git-hunk.git
cd git-hunk
go install ./cmd/git-hunk/...
```

---

## Quick start

```bash
# 1. Make a commit as usual.
git add . && git commit -m "feat: add validation"

# 2. Generate hunk metadata (placeholder summaries without API key).
git hunk generate

# 3. Show metadata for the latest commit.
git hunk show

# 4. Push metadata alongside your branch.
git hunk push

# 5. Verify all hunks have metadata before merging.
git hunk verify
```

---

## Config

Create `~/.config/git-hunk/config.toml`:

```toml
# OpenAI API key for LLM summaries (optional).
api_key = "sk-..."

# Model to use. Defaults to gpt-4o-mini.
model = "gpt-4o-mini"

# Generate metadata automatically on every commit (requires install-hooks).
auto_generate = false
```

---

## Subcommands

| Command | Description |
|---------|-------------|
| `git hunk generate` | Extract hunks from HEAD and write metadata |
| `git hunk show [<rev>]` | Print metadata for a commit (default: HEAD) |
| `git hunk push [<remote>]` | Push `refs/meta/hunks` to remote (default: origin) |
| `git hunk fetch [<remote>]` | Fetch `refs/meta/hunks` from remote (default: origin) |
| `git hunk verify [<rev>]` | Check every hunk in a commit has metadata |
| `git hunk install-hooks` | Install `prepare-commit-msg` and `post-rewrite` hooks |

---

## Storage protocol

Metadata is stored as JSON blobs in a dedicated Git ref (`refs/meta/hunks`).
Each blob is named by the SHA-256 hash of the normalised hunk body, making it
stable across history rewrites.

See [spec/STORAGE.md](spec/STORAGE.md) for the full protocol so other tools
(VS Code extensions, GitKraken plug-ins) can implement rendering.

The JSON schema is at [schema/v1.json](schema/v1.json).

---

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for prerequisites, setup, and the PR
process.

```bash
make test          # run Go tests
make test-cover    # tests + coverage report
make lint          # golangci-lint
make pester        # Pester 5 tests (PowerShell 7 required)
```
