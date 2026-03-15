// Package hunk provides hunk extraction from unified diff output and
// content-addressed hashing so metadata survives rebases and cherry-picks.
package hunk

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Hunk represents a single unified-diff hunk.
type Hunk struct {
	// File is the target (new) filename from the diff header.
	File string
	// OldStart is the line number in the old file.
	OldStart int
	// OldLines is the number of lines from the old file.
	OldLines int
	// NewStart is the line number in the new file.
	NewStart int
	// NewLines is the number of lines from the new file.
	NewLines int
	// Body contains all lines of the hunk (including the @@ header line).
	Body string
}

// Hash returns the stable SHA-256 content hash of the hunk.
// The content is normalised: trailing whitespace is stripped from each line and
// line endings are unified to LF before hashing.
func (h Hunk) Hash() string {
	return HashContent(h.Body)
}

// HashContent normalises s and returns its SHA-256 hex digest.
func HashContent(s string) string {
	normalised := normalise(s)
	sum := sha256.Sum256([]byte(normalised))
	return fmt.Sprintf("%x", sum)
}

// normalise strips trailing whitespace from every line and unifies line endings.
func normalise(s string) string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	var sb strings.Builder
	first := true
	for scanner.Scan() {
		if !first {
			sb.WriteByte('\n')
		}
		first = false
		// Scanner already strips \r from CRLF; only strip trailing spaces/tabs.
		sb.WriteString(strings.TrimRight(scanner.Text(), " \t"))
	}
	return sb.String()
}

// hunkHeaderRe matches a unified diff hunk header, e.g. "@@ -1,4 +1,6 @@".
var hunkHeaderRe = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

// fileHeaderRe matches the +++ line from a diff header.
var fileHeaderRe = regexp.MustCompile(`^\+\+\+ b/(.+)`)

// Extract parses unified diff output from r and returns all hunks found.
func Extract(r io.Reader) ([]Hunk, error) {
	scanner := bufio.NewScanner(r)
	var hunks []Hunk
	var currentFile string
	var currentHunk *Hunk
	var bodyBuf strings.Builder

	flush := func() {
		if currentHunk != nil {
			currentHunk.Body = bodyBuf.String()
			hunks = append(hunks, *currentHunk)
			currentHunk = nil
			bodyBuf.Reset()
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		if m := fileHeaderRe.FindStringSubmatch(line); m != nil {
			flush()
			currentFile = m[1]
			continue
		}

		if m := hunkHeaderRe.FindStringSubmatch(line); m != nil {
			flush()
			h := &Hunk{
				File:     currentFile,
				OldStart: atoi(m[1]),
				OldLines: atoiDefault(m[2], 1),
				NewStart: atoi(m[3]),
				NewLines: atoiDefault(m[4], 1),
			}
			currentHunk = h
			bodyBuf.WriteString(line)
			continue
		}

		if currentHunk != nil {
			bodyBuf.WriteByte('\n')
			bodyBuf.WriteString(line)
		}
	}
	flush()

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return hunks, nil
}

// ExtractFromString is a convenience wrapper around Extract.
func ExtractFromString(diff string) ([]Hunk, error) {
	return Extract(strings.NewReader(diff))
}

func atoi(s string) int {
	if s == "" {
		return 0
	}
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	return atoi(s)
}
