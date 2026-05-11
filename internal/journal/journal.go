package journal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry records one gflow mutating operation.
type Entry struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Op        string            `json:"op"`         // e.g. "feature finish"
	Modules   []string          `json:"modules"`    // module names involved
	RefsBefore map[string]string `json:"refs_before"` // module → SHA before
	RefsAfter  map[string]string `json:"refs_after"`  // module → SHA after
	Results   []ModuleResult    `json:"results"`
}

// ModuleResult records per-module outcome.
type ModuleResult struct {
	Module string `json:"module"`
	Status string `json:"status"` // ok | error | skip
	Error  string `json:"error,omitempty"`
}

// Journal writes and reads operation entries stored in .git/gflow/journal.jsonl.
type Journal struct {
	path string // path to journal.jsonl
}

// Open returns a Journal rooted at gitDir (the .git directory or repo root).
func Open(repoRoot string) (*Journal, error) {
	dir := filepath.Join(repoRoot, ".git", "gflow")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create journal dir: %w", err)
	}
	return &Journal{path: filepath.Join(dir, "journal.jsonl")}, nil
}

// Append writes one entry to the journal.
func (j *Journal) Append(_ context.Context, e Entry) error {
	f, err := os.OpenFile(j.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open journal: %w", err)
	}
	defer f.Close()

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%s\n", b)
	return err
}

// ReadAll returns all entries, oldest first.
func (j *Journal) ReadAll() ([]Entry, error) {
	data, err := os.ReadFile(j.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read journal: %w", err)
	}

	var entries []Entry
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue // corrupt line — skip gracefully
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// Last returns the most recent entry, or nil if journal empty.
func (j *Journal) Last() (*Entry, error) {
	all, err := j.ReadAll()
	if err != nil || len(all) == 0 {
		return nil, err
	}
	e := all[len(all)-1]
	return &e, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
