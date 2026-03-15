package storage_test

import (
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/oscarhfnorris/git-hunk/internal/storage"
)

// ---- helpers ----------------------------------------------------------------

func newInMemoryStore(t *testing.T) *storage.Store {
	t.Helper()
	r, err := gogit.Init(memory.NewStorage(), nil)
	require.NoError(t, err)
	return storage.New(storage.NewGoGitRepository(r))
}

func sampleMetadata(hunkHash string) storage.Metadata {
	return storage.Metadata{
		HunkHash:      hunkHash,
		File:          "pkg/foo/foo.go",
		OldStart:      1,
		OldLines:      5,
		NewStart:      1,
		NewLines:      7,
		Summary:       "Add validation",
		Details:       "Calls validate() before returning.",
		GeneratedBy:   "git-hunk/test",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		SchemaVersion: storage.SchemaVersion,
	}
}

// ---- unit tests -------------------------------------------------------------

func TestStore_Write_Read_RoundTrip(t *testing.T) {
	s := newInMemoryStore(t)
	m := sampleMetadata("abc123")

	require.NoError(t, s.Write(m))

	got, err := s.Read("abc123")
	require.NoError(t, err)

	assert.Equal(t, m.HunkHash, got.HunkHash)
	assert.Equal(t, m.File, got.File)
	assert.Equal(t, m.Summary, got.Summary)
	assert.Equal(t, m.SchemaVersion, got.SchemaVersion)
}

func TestStore_Read_NotFound(t *testing.T) {
	s := newInMemoryStore(t)
	_, err := s.Read("nonexistent")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestStore_Write_EmptyHunkHash(t *testing.T) {
	s := newInMemoryStore(t)
	err := s.Write(storage.Metadata{})
	assert.Error(t, err)
}

func TestStore_Write_SetsSchemaVersion(t *testing.T) {
	s := newInMemoryStore(t)
	m := storage.Metadata{HunkHash: "hash1", File: "f.go"}
	require.NoError(t, s.Write(m))

	got, err := s.Read("hash1")
	require.NoError(t, err)
	assert.Equal(t, storage.SchemaVersion, got.SchemaVersion)
}

func TestStore_Write_SetsGeneratedAt(t *testing.T) {
	s := newInMemoryStore(t)
	m := storage.Metadata{HunkHash: "hash2", File: "f.go"}
	require.NoError(t, s.Write(m))

	got, err := s.Read("hash2")
	require.NoError(t, err)
	assert.NotEmpty(t, got.GeneratedAt)
}

func TestStore_Write_Overwrite(t *testing.T) {
	s := newInMemoryStore(t)

	m1 := sampleMetadata("over1")
	m1.Summary = "first"
	require.NoError(t, s.Write(m1))

	m2 := sampleMetadata("over1")
	m2.Summary = "second"
	require.NoError(t, s.Write(m2))

	got, err := s.Read("over1")
	require.NoError(t, err)
	assert.Equal(t, "second", got.Summary)
}

// ---- suite tests ------------------------------------------------------------

type StoreSuite struct {
	suite.Suite
	store *storage.Store
}

func (s *StoreSuite) SetupTest() {
	r, err := gogit.Init(memory.NewStorage(), nil)
	s.Require().NoError(err)
	s.store = storage.New(storage.NewGoGitRepository(r))
}

func (s *StoreSuite) TestMultipleWrites() {
	for i, hash := range []string{"h1", "h2", "h3"} {
		m := sampleMetadata(hash)
		m.OldStart = i + 1
		s.Require().NoError(s.store.Write(m))
	}
	for i, hash := range []string{"h1", "h2", "h3"} {
		got, err := s.store.Read(hash)
		s.Require().NoError(err)
		s.Equal(i+1, got.OldStart)
	}
}

func (s *StoreSuite) TestMetaRefConstant() {
	s.Equal("refs/meta/hunks", storage.MetaRef)
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}
