package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassifyPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	cases := []struct {
		path string
		want Source
	}{
		{"/opt/homebrew/Cellar/gflow/3.0.0/bin/gflow", SourceHomebrew},
		{"/usr/local/Cellar/gflow/3.0.0/bin/gflow", SourceHomebrew},
		{"/home/linuxbrew/.linuxbrew/bin/gflow", SourceHomebrew},
		{"/usr/local/bin/gflow", SourceBinary},
		{"/usr/bin/gflow", SourceBinary},
		{filepath.Join(home, ".local", "bin", "gflow"), SourceBinary},
		{filepath.Join(home, "bin", "gflow"), SourceBinary},
		{"/tmp/gflow", SourceUnknown},
		{"/home/user/projects/gflow/bin/gflow", SourceUnknown},
	}

	for _, tc := range cases {
		got := classifyPath(tc.path, home)
		if got != tc.want {
			t.Errorf("classifyPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
