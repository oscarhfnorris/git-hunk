// Package storage implements the read/write layer for the refs/meta/hunks Git
// ref. Metadata blobs are stored as a tree of JSON objects, each named by the
// hunk content hash.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// MetaRef is the Git ref under which all hunk metadata is stored.
const MetaRef = "refs/meta/hunks"

// SchemaVersion is the current JSON schema version written into every blob.
const SchemaVersion = "v1"

// Metadata is the JSON structure stored for each hunk.
type Metadata struct {
	HunkHash      string `json:"hunk_hash"`
	File          string `json:"file"`
	OldStart      int    `json:"old_start"`
	OldLines      int    `json:"old_lines"`
	NewStart      int    `json:"new_start"`
	NewLines      int    `json:"new_lines"`
	Summary       string `json:"summary"`
	Details       string `json:"details"`
	GeneratedBy   string `json:"generated_by"`
	GeneratedAt   string `json:"generated_at"`
	SchemaVersion string `json:"schema_version"`
}

// Store manages hunk metadata in a Git repository.
type Store struct {
	repo Repository
}

// Repository is the interface the Store uses to interact with a Git repo.
// Keeping it as an interface allows unit tests to mock it.
type Repository interface {
	// Storer returns the underlying object storer.
	Storer() Storer
}

// Storer is a subset of go-git's storage.Storer used by the Store.
type Storer interface {
	// SetReference persists a reference.
	SetReference(*plumbing.Reference) error
	// Reference retrieves a reference by name.
	Reference(plumbing.ReferenceName) (*plumbing.Reference, error)
	// EncodedObject encodes an object.
	EncodedObject(plumbing.ObjectType, plumbing.Hash) (plumbing.EncodedObject, error)
	// SetEncodedObject writes a raw object.
	SetEncodedObject(plumbing.EncodedObject) (plumbing.Hash, error)
	// NewEncodedObject returns a new mutable encoded object.
	NewEncodedObject() plumbing.EncodedObject
}

// GoGitRepository wraps *gogit.Repository to satisfy the Repository interface.
type GoGitRepository struct {
	r *gogit.Repository
}

// NewGoGitRepository wraps a *gogit.Repository.
func NewGoGitRepository(r *gogit.Repository) *GoGitRepository {
	return &GoGitRepository{r: r}
}

// Storer returns the storer from the underlying go-git repository.
func (g *GoGitRepository) Storer() Storer {
	return &goGitStorer{s: g.r.Storer}
}

// goGitStorer adapts go-git's storer to our Storer interface.
type goGitStorer struct {
	s interface {
		SetReference(*plumbing.Reference) error
		Reference(plumbing.ReferenceName) (*plumbing.Reference, error)
		EncodedObject(plumbing.ObjectType, plumbing.Hash) (plumbing.EncodedObject, error)
		SetEncodedObject(plumbing.EncodedObject) (plumbing.Hash, error)
		NewEncodedObject() plumbing.EncodedObject
	}
}

func (g *goGitStorer) SetReference(r *plumbing.Reference) error {
	return g.s.SetReference(r)
}
func (g *goGitStorer) Reference(n plumbing.ReferenceName) (*plumbing.Reference, error) {
	return g.s.Reference(n)
}
func (g *goGitStorer) EncodedObject(t plumbing.ObjectType, h plumbing.Hash) (plumbing.EncodedObject, error) {
	return g.s.EncodedObject(t, h)
}
func (g *goGitStorer) SetEncodedObject(o plumbing.EncodedObject) (plumbing.Hash, error) {
	return g.s.SetEncodedObject(o)
}
func (g *goGitStorer) NewEncodedObject() plumbing.EncodedObject {
	return g.s.NewEncodedObject()
}

// New creates a new Store backed by the given Repository.
func New(repo Repository) *Store {
	return &Store{repo: repo}
}

// Write serialises m and stores it as a blob in the refs/meta/hunks tree.
// The blob is named by m.HunkHash. The tree and ref are updated atomically.
func (s *Store) Write(m Metadata) error {
	if m.HunkHash == "" {
		return errors.New("storage: HunkHash must not be empty")
	}
	if m.SchemaVersion == "" {
		m.SchemaVersion = SchemaVersion
	}
	if m.GeneratedAt == "" {
		m.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("storage: marshal metadata: %w", err)
	}

	st := s.repo.Storer()

	// Write the blob object.
	blobHash, err := writeBlob(st, data)
	if err != nil {
		return fmt.Errorf("storage: write blob: %w", err)
	}

	// Load or create the tree.
	entries, err := loadTree(st)
	if err != nil {
		return fmt.Errorf("storage: load tree: %w", err)
	}

	// Upsert entry for this hash.
	entries = upsertEntry(entries, m.HunkHash, blobHash)

	// Write new tree.
	treeHash, err := writeTree(st, entries)
	if err != nil {
		return fmt.Errorf("storage: write tree: %w", err)
	}

	// Write a commit wrapping the tree.
	commitHash, err := writeCommit(st, treeHash, m.HunkHash)
	if err != nil {
		return fmt.Errorf("storage: write commit: %w", err)
	}

	// Update the ref.
	ref := plumbing.NewHashReference(plumbing.ReferenceName(MetaRef), commitHash)
	return st.SetReference(ref)
}

// Read retrieves the Metadata stored for hunkHash, or returns
// (zero, ErrNotFound) if no metadata exists for that hash.
func (s *Store) Read(hunkHash string) (Metadata, error) {
	st := s.repo.Storer()

	entries, err := loadTree(st)
	if err != nil {
		return Metadata{}, fmt.Errorf("storage: load tree: %w", err)
	}

	for _, e := range entries {
		if e.Name == hunkHash {
			blob, err := st.EncodedObject(plumbing.BlobObject, e.Hash)
			if err != nil {
				return Metadata{}, fmt.Errorf("storage: read blob: %w", err)
			}
			r, err := blob.Reader()
			if err != nil {
				return Metadata{}, fmt.Errorf("storage: blob reader: %w", err)
			}
			defer r.Close()

			var m Metadata
			if err := json.NewDecoder(r).Decode(&m); err != nil {
				return Metadata{}, fmt.Errorf("storage: decode metadata: %w", err)
			}
			return m, nil
		}
	}
	return Metadata{}, ErrNotFound
}

// ErrNotFound is returned when no metadata blob exists for a given hunk hash.
var ErrNotFound = errors.New("storage: metadata not found")

// ---- helpers ----------------------------------------------------------------

func writeBlob(st Storer, data []byte) (plumbing.Hash, error) {
	obj := st.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	obj.SetSize(int64(len(data)))
	w, err := obj.Writer()
	if err != nil {
		return plumbing.ZeroHash, err
	}
	if _, err = w.Write(data); err != nil {
		return plumbing.ZeroHash, err
	}
	if err = w.Close(); err != nil {
		return plumbing.ZeroHash, err
	}
	return st.SetEncodedObject(obj)
}

func writeTree(st Storer, entries []object.TreeEntry) (plumbing.Hash, error) {
	tree := &object.Tree{Entries: entries}
	obj := st.NewEncodedObject()
	if err := tree.Encode(obj); err != nil {
		return plumbing.ZeroHash, err
	}
	return st.SetEncodedObject(obj)
}

func writeCommit(st Storer, treeHash plumbing.Hash, msg string) (plumbing.Hash, error) {
	// Look up current ref to use as parent.
	var parents []plumbing.Hash
	ref, err := st.Reference(plumbing.ReferenceName(MetaRef))
	if err == nil {
		parents = []plumbing.Hash{ref.Hash()}
	}

	sig := object.Signature{
		Name:  "git-hunk",
		Email: "git-hunk@localhost",
		When:  time.Now().UTC(),
	}
	commit := &object.Commit{
		Author:    sig,
		Committer: sig,
		Message:   "chore: update hunk metadata for " + msg,
		TreeHash:  treeHash,
		ParentHashes: parents,
	}
	obj := st.NewEncodedObject()
	if err := commit.Encode(obj); err != nil {
		return plumbing.ZeroHash, err
	}
	return st.SetEncodedObject(obj)
}

func loadTree(st Storer) ([]object.TreeEntry, error) {
	ref, err := st.Reference(plumbing.ReferenceName(MetaRef))
	if err != nil {
		// No ref yet — start with empty tree.
		return nil, nil //nolint:nilerr
	}

	commitObj, err := st.EncodedObject(plumbing.CommitObject, ref.Hash())
	if err != nil {
		return nil, err
	}
	commitR, err := commitObj.Reader()
	if err != nil {
		return nil, err
	}
	defer commitR.Close()

	var commit object.Commit
	if err := commit.Decode(commitObj); err != nil {
		return nil, err
	}

	treeObj, err := st.EncodedObject(plumbing.TreeObject, commit.TreeHash)
	if err != nil {
		return nil, err
	}
	var tree object.Tree
	if err := tree.Decode(treeObj); err != nil {
		return nil, err
	}
	return tree.Entries, nil
}

func upsertEntry(entries []object.TreeEntry, name string, hash plumbing.Hash) []object.TreeEntry {
	for i, e := range entries {
		if e.Name == name {
			entries[i].Hash = hash
			return entries
		}
	}
	return append(entries, object.TreeEntry{
		Name: name,
		Mode: filemode.Regular,
		Hash: hash,
	})
}
