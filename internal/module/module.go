package module

// Module represents a single git repository (root or submodule) in the monorepo.
type Module struct {
	Name string // display name
	Path string // absolute path on disk
	Root bool   // true if this is the superproject root
}
