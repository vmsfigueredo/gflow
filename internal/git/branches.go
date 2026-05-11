package git

import (
	"context"
	"strings"
	"sync"

	"github.com/vmsfigueredo/gflow/internal/module"
)

// DetectCurrentFlowBranch returns the branch suffix if the module is on a
// branch that starts with prefix (e.g. "feature/"). Returns ("", false, nil)
// when not on a flow branch.
func DetectCurrentFlowBranch(ctx context.Context, dir, prefix string) (name string, ok bool, err error) {
	b, err := CurrentBranch(ctx, dir)
	if err != nil {
		return "", false, err
	}
	b = strings.TrimSpace(b)
	if !strings.HasPrefix(b, prefix) {
		return "", false, nil
	}
	return strings.TrimPrefix(b, prefix), true, nil
}

// BatchCurrentBranches concurrently fetches the current branch for every
// module, returning a map from module.Name to branch string.
// Failures are silently swallowed — callers treat missing entries as unknown.
func BatchCurrentBranches(ctx context.Context, mods []*module.Module) map[string]string {
	out := make(map[string]string, len(mods))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Bound concurrency to 8 to keep startup latency low.
	sem := make(chan struct{}, 8)

	for _, m := range mods {
		m := m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			b, err := CurrentBranch(ctx, m.Path)
			if err == nil {
				mu.Lock()
				out[m.Name] = strings.TrimSpace(b)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return out
}
