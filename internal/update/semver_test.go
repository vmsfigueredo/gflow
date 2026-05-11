package update

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v3.0.1", "v3.0.0", true},
		{"v3.0.0", "v3.0.0", false},
		{"v3.0.0", "v3.0.1", false},
		{"v3.0.1", "dev", true}, // dev always outdated
		{"v4.0.0", "v3.9.9", true},
	}

	for _, tc := range cases {
		got, err := IsNewer(tc.latest, tc.current)
		if err != nil {
			t.Errorf("IsNewer(%q, %q) error: %v", tc.latest, tc.current, err)
			continue
		}
		if got != tc.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}
