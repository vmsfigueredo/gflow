package module

// Module represents a single git repository (root or submodule) in the monorepo.
type Module struct {
	Name    string // identifier (filepath.Base or ".") — used as map key and path component
	Display string // human-facing label: alias or parent/dir
	Path    string // absolute path on disk
	Root    bool   // true if this is the superproject root
}
