package git

import (
	"context"
	"strings"
	"sync"
)

type lsRemoteCache struct {
	mu    sync.Mutex
	store map[string]string // "dir|remote|ref" → SHA or ""
}

var globalLsRemoteCache = &lsRemoteCache{store: make(map[string]string)}

// LsRemoteRef returns the SHA of a remote ref, or "" if not found.
// Results are memoized for the process lifetime (fixes detect_main_branch perf gap).
func LsRemoteRef(ctx context.Context, dir, remote, ref string) (string, error) {
	key := dir + "|" + remote + "|" + ref

	globalLsRemoteCache.mu.Lock()
	if v, ok := globalLsRemoteCache.store[key]; ok {
		globalLsRemoteCache.mu.Unlock()
		return v, nil
	}
	globalLsRemoteCache.mu.Unlock()

	res, err := Run(ctx, dir, "ls-remote", "--heads", remote, ref)
	if err != nil {
		return "", err
	}

	sha := ""
	if fields := strings.Fields(res.Stdout); len(fields) > 0 {
		sha = fields[0]
	}

	globalLsRemoteCache.mu.Lock()
	globalLsRemoteCache.store[key] = sha
	globalLsRemoteCache.mu.Unlock()

	return sha, nil
}

// DetectMainBranch returns "main" if it exists on remote, else "master".
func DetectMainBranch(ctx context.Context, dir, remote string) (string, error) {
	for _, candidate := range []string{"main", "master"} {
		sha, err := LsRemoteRef(ctx, dir, remote, candidate)
		if err != nil {
			return "", err
		}
		if sha != "" {
			return candidate, nil
		}
	}
	return "main", nil
}

// IsRemoteSynced returns true if local branch is up-to-date with remote.
func IsRemoteSynced(ctx context.Context, dir, remote, branch string) (bool, error) {
	localRes, err := Run(ctx, dir, "rev-parse", branch)
	if err != nil {
		return false, err
	}
	remoteRef := remote + "/" + branch
	remoteRes, err := Run(ctx, dir, "rev-parse", remoteRef)
	if err != nil {
		return false, err
	}
	return localRes.Stdout == remoteRes.Stdout, nil
}
