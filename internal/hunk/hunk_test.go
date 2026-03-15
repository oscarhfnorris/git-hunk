package hunk_test

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/oscarhfnorris/git-hunk/internal/hunk"
)

// ---- table-driven hash tests ----

func TestHashContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		// We only verify that identical normalised content produces the same hash
		// and different content produces different hashes; we do NOT hard-code
		// the exact digest so the tests remain readable.
		sameAs string // if non-empty, expect same hash as this input
	}{
		{
			name:  "trailing spaces stripped",
			input: "line one   \nline two\t\n",
			sameAs: "line one\nline two\n",
		},
		{
			name:  "CRLF normalised",
			input: "line one\r\nline two\r\n",
			sameAs: "line one\nline two\n",
		},
		{
			name:  "empty string",
			input: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hunk.HashContent(tc.input)
			assert.NotEmpty(t, got, "hash must not be empty")
			assert.Len(t, got, 64, "SHA-256 hex must be 64 chars")
			if tc.sameAs != "" {
				want := hunk.HashContent(tc.sameAs)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("hash mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestHashContent_DifferentInputs(t *testing.T) {
	h1 := hunk.HashContent("alpha")
	h2 := hunk.HashContent("beta")
	assert.NotEqual(t, h1, h2)
}

// ---- extract tests using testdata fixtures ----

func TestExtract_SimpleFixture(t *testing.T) {
	f, err := os.Open("../../testdata/simple.patch")
	require.NoError(t, err)
	defer f.Close()

	hunks, err := hunk.Extract(f)
	require.NoError(t, err)
	require.Len(t, hunks, 2)

	assert.Equal(t, "main.go", hunks[0].File)
	assert.Equal(t, 1, hunks[0].OldStart)
	assert.Equal(t, 5, hunks[0].OldLines)
	assert.Equal(t, 1, hunks[0].NewStart)
	assert.Equal(t, 7, hunks[0].NewLines)

	assert.Equal(t, "main.go", hunks[1].File)
}

func TestExtract_MultiFile(t *testing.T) {
	f, err := os.Open("../../testdata/multi_file.patch")
	require.NoError(t, err)
	defer f.Close()

	hunks, err := hunk.Extract(f)
	require.NoError(t, err)
	require.Len(t, hunks, 2)

	assert.Equal(t, "pkg/foo/foo.go", hunks[0].File)
	assert.Equal(t, "pkg/bar/bar.go", hunks[1].File)
}

func TestExtract_EmptyInput(t *testing.T) {
	hunks, err := hunk.ExtractFromString("")
	require.NoError(t, err)
	assert.Empty(t, hunks)
}

func TestHunk_Hash_Stable(t *testing.T) {
	diff := `@@ -1,3 +1,4 @@
 package foo
+// comment
 func Foo() {}`
	hunks, err := hunk.ExtractFromString(diff)
	require.NoError(t, err)
	require.Len(t, hunks, 1)

	h1 := hunks[0].Hash()
	h2 := hunks[0].Hash()
	assert.Equal(t, h1, h2, "hash must be deterministic")
}

// ---- suite tests ----

type HunkSuite struct {
	suite.Suite
	simplePatch string
}

func (s *HunkSuite) SetupSuite() {
	data, err := os.ReadFile("../../testdata/simple.patch")
	s.Require().NoError(err)
	s.simplePatch = string(data)
}

func (s *HunkSuite) TestExtractFromString_ReturnsHunks() {
	hunks, err := hunk.ExtractFromString(s.simplePatch)
	s.Require().NoError(err)
	s.Len(hunks, 2)
}

func (s *HunkSuite) TestHunkBodyNotEmpty() {
	hunks, err := hunk.ExtractFromString(s.simplePatch)
	s.Require().NoError(err)
	for _, h := range hunks {
		s.NotEmpty(h.Body)
	}
}

func (s *HunkSuite) TestHashLength() {
	hunks, err := hunk.ExtractFromString(s.simplePatch)
	s.Require().NoError(err)
	for _, h := range hunks {
		s.Len(h.Hash(), 64)
	}
}

func (s *HunkSuite) TestExtractFromString_NoHunks() {
	hunks, err := hunk.ExtractFromString("not a diff")
	s.Require().NoError(err)
	s.Empty(hunks)
}

func (s *HunkSuite) TestFile_SetFromPlusPlusPlus() {
	diff := `+++ b/some/file.go
@@ -1,2 +1,3 @@
 line1
+line2
 line3`
	hunks, err := hunk.ExtractFromString(diff)
	s.Require().NoError(err)
	s.Require().Len(hunks, 1)
	s.Equal("some/file.go", hunks[0].File)
}

func TestHunkSuite(t *testing.T) {
	suite.Run(t, new(HunkSuite))
}

// ---- additional normalisation tests ----

func TestHashContent_NormalisedEquivalence(t *testing.T) {
	pairs := []struct {
		a, b string
	}{
		{"hello \nworld", "hello\nworld"},
		{"foo\r\nbar\r\n", "foo\nbar\n"},
		{"  trailing  \n  spaces  \n", "  trailing\n  spaces\n"},
	}
	for _, p := range pairs {
		ha := hunk.HashContent(p.a)
		hb := hunk.HashContent(p.b)
		if diff := cmp.Diff(ha, hb); diff != "" {
			t.Errorf("expected equal hashes for %q and %q, diff: %s", p.a, p.b, diff)
		}
	}
}

func TestExtract_HunkBodyContainsHeader(t *testing.T) {
	diff := "@@ -1,2 +1,3 @@\n line\n+added\n line2"
	hunks, err := hunk.ExtractFromString(diff)
	require.NoError(t, err)
	require.Len(t, hunks, 1)
	assert.True(t, strings.HasPrefix(hunks[0].Body, "@@ "))
}
