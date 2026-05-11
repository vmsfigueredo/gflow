package gitflow

import (
	"context"
	"strings"

	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// RunOpFn is the function used to execute a git-flow operation.
// Overridable in tests.
var RunOpFn = RunOp

// RunOp dispatches a git-flow operation for one module.
// Passes real FinishOptions from config on finish.
func RunOp(ctx context.Context, variant Variant, cfg *config.Config, m *module.Module,
	branchType, op, name string, dryRun bool) (string, error) {

	if dryRun {
		return "git flow " + branchType + " " + op + " " + name, nil
	}

	args := buildArgs(variant, cfg, branchType, op, name)
	return runGitFlow(ctx, m.Path, args...)
}

func buildArgs(variant Variant, cfg *config.Config, branchType, op, name string) []string {
	args := []string{branchType, op, name}

	if op == "finish" && cfg.Gitflow.FinishOptions != "" {
		opts := strings.Fields(cfg.Gitflow.FinishOptions)
		args = append(args, opts...)
	}

	// variant-specific flags for delete/finish
	if op == "delete" {
		switch variant {
		case VariantAVH:
			args = append(args, "--remote")
		}
	}

	return args
}
