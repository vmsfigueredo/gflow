package executor

import (
	"context"

	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// OpFunc is a unit of work for one module.
type OpFunc func(ctx interface{}, m *module.Module) Result

// Runner dispatches operations across modules.
type Runner interface {
	Run(ctx context.Context, mods []*module.Module, op OpFunc) []Result
}

// New returns a parallel or serial runner based on flags.
func New(cfg *config.Config, parallel, failFast, dryRun, debug bool) Runner {
	if parallel || cfg.Parallel {
		return &parallelRunner{
			workers:  cfg.Workers,
			failFast: failFast,
			dryRun:   dryRun,
			debug:    debug,
		}
	}
	return &serialRunner{failFast: failFast, dryRun: dryRun, debug: debug}
}
