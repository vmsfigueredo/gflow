package executor

import (
	"context"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/vmsfigueredo/gflow/internal/module"
)

type parallelRunner struct {
	workers  int
	failFast bool
	dryRun   bool
	debug    bool
}

func (r *parallelRunner) Run(ctx context.Context, mods []*module.Module, op OpFunc) []Result {
	workers := r.workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > len(mods) {
		workers = len(mods)
	}

	results := make([]Result, len(mods))

	// per-.git-path mutex prevents index.lock contention when modules share a worktree
	var gitLockMu sync.Map

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(workers)

	for i, m := range mods {
		i, m := i, m
		eg.Go(func() error {
			gitDir := filepath.Join(m.Path, ".git")
			muVal, _ := gitLockMu.LoadOrStore(gitDir, &sync.Mutex{})
			mu := muVal.(*sync.Mutex)
			mu.Lock()
			defer mu.Unlock()

			start := time.Now()
			res := op(ctx, m)
			res.Duration = time.Since(start)
			results[i] = res

			if r.failFast && res.Status == StatusError {
				return res.Err
			}
			return nil
		})
	}

	_ = eg.Wait()
	return results
}
