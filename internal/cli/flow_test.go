package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNamesFlag(t *testing.T) {
	cases := []struct {
		in   string
		want map[string]string
	}{
		{"", nil},
		{"   ", nil},
		{"api=v1.2.3", map[string]string{"api": "v1.2.3"}},
		{"api=v1.2.3,web=v1.4.0", map[string]string{"api": "v1.2.3", "web": "v1.4.0"}},
		{"api=v1.2.3, web=v1.4.0", map[string]string{"api": "v1.2.3", "web": "v1.4.0"}},
		// malformed pairs are silently dropped
		{"noequals,api=good", map[string]string{"api": "good"}},
		{"=nokey,api=good", map[string]string{"api": "good"}},
		{"api=", nil}, // empty value → dropped
	}
	for _, c := range cases {
		got := parseNamesFlag(c.in)
		assert.Equal(t, c.want, got, "input: %q", c.in)
	}
}
