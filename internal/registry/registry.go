package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// Entry is a registered project.
type Entry struct {
	Alias    string    `yaml:"alias"`
	Path     string    `yaml:"path"`
	LastUsed time.Time `yaml:"last_used,omitempty"`
}

// Registry manages ~/.config/gflow/projects.yaml.
type Registry struct {
	path    string
	Entries []Entry `yaml:"projects"`
}

// Open loads the global registry (creates file + dirs if missing).
func Open() (*Registry, error) {
	path, err := defaultPath()
	if err != nil {
		return nil, err
	}
	r := &Registry{path: path}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

// Add registers a new project. Returns error if alias already exists.
func (r *Registry) Add(alias, absPath string) error {
	for _, e := range r.Entries {
		if e.Alias == alias {
			return fmt.Errorf("alias %q already registered (path: %s)", alias, e.Path)
		}
	}
	r.Entries = append(r.Entries, Entry{
		Alias:    alias,
		Path:     absPath,
		LastUsed: time.Now(),
	})
	return r.save()
}

// Remove deletes an entry by alias.
func (r *Registry) Remove(alias string) error {
	for i, e := range r.Entries {
		if e.Alias == alias {
			r.Entries = append(r.Entries[:i], r.Entries[i+1:]...)
			return r.save()
		}
	}
	return fmt.Errorf("alias %q not found", alias)
}

// Get returns the entry for alias, or nil.
func (r *Registry) Get(alias string) *Entry {
	for i, e := range r.Entries {
		if e.Alias == alias {
			return &r.Entries[i]
		}
	}
	return nil
}

// Recent returns entries sorted by LastUsed descending.
func (r *Registry) Recent(n int) []Entry {
	sorted := make([]Entry, len(r.Entries))
	copy(sorted, r.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastUsed.After(sorted[j].LastUsed)
	})
	if n > 0 && n < len(sorted) {
		return sorted[:n]
	}
	return sorted
}

// Touch updates LastUsed for the entry matching absPath, if found.
func (r *Registry) Touch(absPath string) error {
	for i, e := range r.Entries {
		if e.Path == absPath {
			r.Entries[i].LastUsed = time.Now()
			return r.save()
		}
	}
	return nil // not registered — silently skip
}

// TouchByContext calls Touch with the current working directory.
func TouchByContext(_ context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return nil // best-effort
	}
	r, err := Open()
	if err != nil {
		return nil
	}
	return r.Touch(cwd)
}

func (r *Registry) load() error {
	data, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read registry: %w", err)
	}
	return yaml.Unmarshal(data, r)
}

func (r *Registry) save() error {
	data, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0o644)
}

func defaultPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	dir := filepath.Join(base, "gflow")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}
	return filepath.Join(dir, "projects.yaml"), nil
}
