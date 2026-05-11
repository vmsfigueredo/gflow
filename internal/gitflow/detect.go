package gitflow

import (
	"context"
	"strings"
)

// Variant identifies the git-flow CLI variant.
type Variant string

const (
	VariantAVH   Variant = "avh"
	VariantNvie  Variant = "nvie"
	VariantUnknown Variant = "unknown"
)

// DetectVariant parses `git flow version` to identify AVH or nvie.
func DetectVariant(ctx context.Context) (Variant, error) {
	res, err := runGitFlow(ctx, "", "version")
	if err != nil {
		return VariantUnknown, err
	}
	out := strings.ToLower(res)
	switch {
	case strings.Contains(out, "avh"):
		return VariantAVH, nil
	default:
		return VariantNvie, nil
	}
}
