package git

import (
	"context"
	"fmt"
	"strings"
)

// SubmoduleAutoCommit stages the submodule pointer for module at subPath in parentDir
// and commits with a message derived from commitMsgTemplate (replaces {module}, {sha}).
func SubmoduleAutoCommit(ctx context.Context, parentDir, subPath, moduleName, msgTemplate string) error {
	sha, err := CurrentSHA(ctx, subPath)
	if err != nil {
		return fmt.Errorf("get sha for %s: %w", moduleName, err)
	}

	msg := strings.ReplaceAll(msgTemplate, "{module}", moduleName)
	msg = strings.ReplaceAll(msg, "{sha}", sha[:8])

	if _, err := Run(ctx, parentDir, "add", subPath); err != nil {
		return fmt.Errorf("git add %s: %w", subPath, err)
	}

	// check if there's something staged (submodule pointer may not have changed)
	res, err := Run(ctx, parentDir, "diff", "--cached", "--name-only")
	if err != nil {
		return err
	}
	if strings.TrimSpace(res.Stdout) == "" {
		return nil // nothing to commit
	}

	if _, err := Run(ctx, parentDir, "commit", "-m", msg); err != nil {
		return fmt.Errorf("git commit pointer for %s: %w", moduleName, err)
	}
	return nil
}

// UpdateInit runs git submodule update --init --recursive.
func UpdateInit(ctx context.Context, dir string) error {
	_, err := Run(ctx, dir, "submodule", "update", "--init", "--recursive")
	return err
}
