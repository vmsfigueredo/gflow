package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PRInfo holds summary data for one pull request.
type PRInfo struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	URL       string `json:"url"`
	HeadRef   string `json:"headRefName"`
	BaseRef   string `json:"baseRefName"`
	Mergeable string `json:"mergeable"`
	ChecksOK  bool
}

// IsAvailable returns true if the gh CLI is on PATH.
func IsAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// PRCreate creates a PR for the current branch in repoDir.
func PRCreate(ctx context.Context, repoDir, title, body, base string, draft bool) (*PRInfo, error) {
	args := []string{"pr", "create",
		"--title", title,
		"--body", body,
		"--base", base,
	}
	if draft {
		args = append(args, "--draft")
	}
	if _, err := runGH(ctx, repoDir, args...); err != nil {
		return nil, err
	}
	return PRStatus(ctx, repoDir)
}

// PRStatus returns PR info for the current branch in repoDir.
func PRStatus(ctx context.Context, repoDir string) (*PRInfo, error) {
	out, err := runGH(ctx, repoDir,
		"pr", "view",
		"--json", "number,title,state,url,headRefName,baseRefName,mergeable",
	)
	if err != nil {
		return nil, err
	}
	var info PRInfo
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		return nil, fmt.Errorf("parse gh pr view output: %w", err)
	}
	return &info, nil
}

// PRMerge merges the PR for the current branch in repoDir using strategy.
// strategy: "squash" | "rebase" | "merge"
func PRMerge(ctx context.Context, repoDir, strategy string) error {
	args := []string{"pr", "merge", "--auto"}
	switch strategy {
	case "squash":
		args = append(args, "--squash")
	case "rebase":
		args = append(args, "--rebase")
	default:
		args = append(args, "--merge")
	}
	_, err := runGH(ctx, repoDir, args...)
	return err
}

func runGH(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}
