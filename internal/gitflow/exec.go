package gitflow

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Verbose streams raw git flow stdout/stderr to os.Stderr when true.
var Verbose bool

// runGitFlow executes `git flow <args>` in dir and returns stdout.
func runGitFlow(ctx context.Context, dir string, args ...string) (string, error) {
	full := append([]string{"flow"}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdout, stderr bytes.Buffer
	if Verbose {
		fmt.Fprintf(os.Stderr, "+ git flow %s (in %s)\n", strings.Join(args, " "), dir)
		cmd.Stdout = io.MultiWriter(&stdout, os.Stderr)
		cmd.Stderr = io.MultiWriter(&stderr, os.Stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git flow %s: exit %d: %s",
				strings.Join(args, " "), exit.ExitCode(), strings.TrimSpace(stderr.String()))
		}
		return "", fmt.Errorf("git flow %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimRight(stdout.String(), "\n"), nil
}
