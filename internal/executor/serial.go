package executor

import (
	"context"
	"time"

	"github.com/vmsfigueredo/gflow/internal/module"
)

type serialRunner struct {
	failFast bool
	dryRun   bool
	debug    bool
}

func (r *serialRunner) Run(ctx context.Context, mods []*module.Module, op OpFunc) []Result {
	results := make([]Result, 0, len(mods))
	for _, m := range mods {
		start := time.Now()
		res := op(ctx, m)
		res.Duration = time.Since(start)
		results = append(results, res)
		if r.failFast && res.Status == StatusError {
			break
		}
	}
	return results
}
