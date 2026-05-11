package git

import (
	"context"
	"strings"
	"sync"

	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// Merge runs git merge with the given args in the module's directory.
func Merge(m *module.Module, args []string, dryRun bool) executor.Result {
	full := append([]string{"merge"}, args...)
	if dryRun {
		return executor.Result{Module: m, Action: "merge", Status: executor.StatusDryRun, Output: "git " + strings.Join(full, " ")}
	}
	_, err := Run(context.Background(), m.Path, full...)
	if err != nil {
		return executor.Result{Module: m, Action: "merge", Status: executor.StatusError, Err: err}
	}
	return executor.Result{Module: m, Action: "merge", Status: executor.StatusOK}
}

// ListLocalBranches returns all local branch names in dir, deduplicated, in git's default order.
func ListLocalBranches(ctx context.Context, dir string) ([]string, error) {
	res, err := Run(ctx, dir, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(res.Stdout)
	if raw == "" {
		return nil, nil
	}
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		if b := strings.TrimSpace(l); b != "" {
			out = append(out, b)
		}
	}
	return out, nil
}

// BatchLocalBranches concurrently collects all local branches across modules and returns
// a deduplicated, ordered slice (order is deterministic: sorted by first appearance across modules).
func BatchLocalBranches(ctx context.Context, mods []*module.Module) []string {
	type result struct {
		branches []string
		idx      int
	}
	results := make([]result, len(mods))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	for i, m := range mods {
		i, m := i, m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			branches, _ := ListLocalBranches(ctx, m.Path)
			results[i] = result{branches: branches, idx: i}
		}()
	}
	wg.Wait()

	seen := make(map[string]bool)
	var out []string
	for _, r := range results {
		for _, b := range r.branches {
			if !seen[b] {
				seen[b] = true
				out = append(out, b)
			}
		}
	}
	return out
}
