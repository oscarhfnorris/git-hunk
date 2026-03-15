# git-hunk Storage Protocol — v1

> This document describes how git-hunk stores per-hunk metadata so that
> other tools (VS Code extensions, GitKraken plug-ins, CI checks) can read
> and write the same data without depending on the git-hunk binary.

---

## 1  Overview

All metadata lives inside the repository itself, under a dedicated Git ref:

```
refs/meta/hunks
```

This ref points to a **commit** whose **tree** contains one blob per hunk.
Each blob is named after the hunk's content hash and contains a JSON object
that conforms to [`schema/v1.json`](../schema/v1.json).

Because the ref is a normal Git ref it can be pushed and fetched like any
other branch:

```bash
git push origin refs/meta/hunks:refs/meta/hunks
git fetch origin refs/meta/hunks:refs/meta/hunks
```

---

## 2  Content-Addressed Key

The blob name (and the `hunk_hash` field inside the JSON) is the **SHA-256
hex digest** of the *normalised* hunk body.

### 2.1  Normalisation rules

Before hashing, apply the following transforms in order:

| Step | Rule |
|------|------|
| 1 | Unify line endings to `\n` (strip `\r`) |
| 2 | Strip trailing whitespace (spaces and tabs) from every line |
| 3 | Do **not** strip leading whitespace |
| 4 | Do **not** add or remove a trailing newline |

### 2.2  Stability guarantee

A hunk that is rebased, cherry-picked, or amended onto a different base
commit will produce the **same hash** as long as its body is unchanged.
This means metadata survives the rewrite operations that are common in
feature-branch workflows.

---

## 3  Tree layout

```
refs/meta/hunks  →  commit
                        └── tree/
                               ├── <hash-1>   (blob, JSON)
                               ├── <hash-2>   (blob, JSON)
                               └── ...
```

All blobs are stored at the root of the tree (no subdirectories).  File
mode for each blob is `100644` (regular file).

---

## 4  Blob format

Each blob is a UTF-8 JSON object.  The canonical schema is defined in
[`schema/v1.json`](../schema/v1.json).  The required fields are:

| Field | Type | Description |
|-------|------|-------------|
| `hunk_hash` | string | 64-char lowercase hex SHA-256 |
| `file` | string | Repo-relative path of the modified file |
| `old_start` | integer | First old-file line number |
| `old_lines` | integer | Number of old-file lines in hunk |
| `new_start` | integer | First new-file line number |
| `new_lines` | integer | Number of new-file lines in hunk |
| `summary` | string | ≤80-char one-line description |
| `details` | string | Extended Markdown explanation (may be empty) |
| `generated_by` | string | Tool identifier, e.g. `"git-hunk/v0.1.0"` |
| `generated_at` | string | RFC 3339 timestamp |
| `schema_version` | string | Always `"v1"` for this revision |

Unknown fields should be preserved by compliant readers (forward
compatibility).

---

## 5  Commit format

Each write to `refs/meta/hunks` creates a new commit with:

* **Author / Committer**: `git-hunk <git-hunk@localhost>`
* **Message**: `chore: update hunk metadata for <hunk_hash[:8]>`
* **Parent**: the previous `refs/meta/hunks` commit (if any)
* **Tree**: the updated blob tree

This gives a linear history of all metadata changes.

---

## 6  Interoperability

Any tool that can read Git objects can consume this data without the
git-hunk binary.  Example using `git cat-file`:

```bash
# List all hunk hashes
git ls-tree refs/meta/hunks

# Read metadata for a specific hash
git cat-file blob refs/meta/hunks:<hunk_hash>
```

---

## 7  Versioning

This is schema version **v1**.  Future breaking changes will use a new
`schema_version` value and a separate section in this document.
