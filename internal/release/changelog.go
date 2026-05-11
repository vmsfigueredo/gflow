package release

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vmsfigueredo/gflow/internal/git"
)

// GenerateChangelog generates a changelog entry for version from commits since lastTag.
// Appends the block at the top of CHANGELOG.md (below the first # Changelog line).
func GenerateChangelog(ctx context.Context, repoDir, version, lastTag string) error {
	log, err := commitsSince(ctx, repoDir, lastTag)
	if err != nil {
		return fmt.Errorf("read commits for changelog: %w", err)
	}

	block := formatBlock(version, log)
	return prependToChangelog(repoDir, block)
}

type commit struct {
	Hash    string
	Subject string
}

func commitsSince(ctx context.Context, repoDir, since string) ([]commit, error) {
	rangeArg := "HEAD"
	if since != "" {
		rangeArg = since + "..HEAD"
	}
	res, err := git.Run(ctx, repoDir, "log", rangeArg, "--pretty=format:%H\t%s", "--no-merges")
	if err != nil {
		return nil, err
	}
	var commits []commit
	for _, line := range strings.Split(res.Stdout, "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 && parts[0] != "" {
			commits = append(commits, commit{Hash: parts[0][:8], Subject: parts[1]})
		}
	}
	return commits, nil
}

func formatBlock(version string, commits []commit) string {
	sections := map[string][]string{
		"feat":     {},
		"fix":      {},
		"refactor": {},
		"perf":     {},
		"chore":    {},
		"other":    {},
	}
	order := []string{"feat", "fix", "refactor", "perf", "chore", "other"}
	sectionLabels := map[string]string{
		"feat":     "### Added",
		"fix":      "### Fixed",
		"refactor": "### Refactored",
		"perf":     "### Performance",
		"chore":    "### Chores",
		"other":    "### Other",
	}

	for _, c := range commits {
		prefix := conventionalPrefix(c.Subject)
		if _, ok := sections[prefix]; !ok {
			prefix = "other"
		}
		sections[prefix] = append(sections[prefix], fmt.Sprintf("- %s (`%s`)", c.Subject, c.Hash))
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "## [%s] — %s\n\n", version, time.Now().Format("2006-01-02"))
	for _, key := range order {
		if len(sections[key]) == 0 {
			continue
		}
		fmt.Fprintf(&b, "%s\n", sectionLabels[key])
		for _, line := range sections[key] {
			fmt.Fprintf(&b, "%s\n", line)
		}
		fmt.Fprintln(&b)
	}
	return b.String()
}

func conventionalPrefix(subject string) string {
	for _, key := range []string{"feat", "fix", "refactor", "perf", "chore"} {
		if strings.HasPrefix(subject, key+"(") || strings.HasPrefix(subject, key+":") {
			return key
		}
	}
	return "other"
}

func prependToChangelog(repoDir, block string) error {
	path := filepath.Join(repoDir, "CHANGELOG.md")
	existing, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		existing = []byte("# Changelog\n\n")
	} else if err != nil {
		return err
	}

	content := string(existing)
	insertAt := 0
	// Insert after first line if it starts with # Changelog
	if lines := strings.SplitN(content, "\n", 3); len(lines) >= 2 {
		insertAt = len(lines[0]) + 1 // after first newline
		if len(lines) >= 2 {
			insertAt += len(lines[1]) + 1 // after blank line
		}
	}

	var out bytes.Buffer
	out.WriteString(content[:insertAt])
	out.WriteString("\n")
	out.WriteString(block)
	out.WriteString(content[insertAt:])

	return os.WriteFile(path, out.Bytes(), 0o644)
}
