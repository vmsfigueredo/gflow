package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// RunResult holds the output of a git subprocess.
type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run executes git with args in dir.
// Returns a typed RunResult — never swallows stderr or reads $? of a pipe.
func Run(ctx context.Context, dir string, args ...string) (RunResult, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := RunResult{
		Stdout: strings.TrimRight(stdout.String(), "\n"),
		Stderr: strings.TrimRight(stderr.String(), "\n"),
	}

	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exit.ExitCode()
			return res, fmt.Errorf("git %s: exit %d: %s", strings.Join(args, " "), res.ExitCode, res.Stderr)
		}
		return res, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return res, nil
}

// MustRun returns stdout or panics — for internal use in test helpers only.
func MustRun(ctx context.Context, dir string, args ...string) string {
	res, err := Run(ctx, dir, args...)
	if err != nil {
		panic(err)
	}
	return res.Stdout
}
