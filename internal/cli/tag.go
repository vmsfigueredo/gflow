package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage git tags (list, push)",
	}
	cmd.AddCommand(
		newTagListCmd(),
		newTagPushCmd(),
	)
	return cmd
}

func newTagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tags in the current repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			res, err := git.Run(cmd.Context(), root, "tag", "--sort=-version:refname")
			if err != nil {
				return err
			}
			printTagList(strings.Fields(strings.TrimSpace(res.Stdout)))
			return nil
		},
	}
}

func printTagList(tags []string) {
	if len(tags) == 0 {
		output.Infof("No tags found.")
		return
	}

	// group by vMAJOR.MINOR
	type group struct {
		key  string
		tags []string
	}
	var groups []group
	index := map[string]int{}

	for _, t := range tags {
		key := minorKey(t)
		if i, ok := index[key]; ok {
			groups[i].tags = append(groups[i].tags, t)
		} else {
			index[key] = len(groups)
			groups = append(groups, group{key: key, tags: []string{t}})
		}
	}

	latest := tags[0]
	fmt.Printf("\n  Latest  %s\n\n", output.HelpInlineCode(latest))

	for _, g := range groups {
		fmt.Printf("  %s\n", output.HelpInlineCode(g.key))
		for _, t := range g.tags {
			line := fmt.Sprintf("    %s", t)
			if t == latest {
				line += "  " + output.HelpUsage("← latest")
			}
			fmt.Println(line)
		}
		fmt.Println()
	}
}

// minorKey returns "vMAJOR.MINOR" for a semver tag, or the tag itself.
func minorKey(tag string) string {
	s := strings.TrimPrefix(tag, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) >= 2 {
		return "v" + parts[0] + "." + parts[1]
	}
	return tag
}

func newTagPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push [tag]",
		Short: "Push tags to remote (all tags if no arg)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			remote := flagRemote

			var gitArgs []string
			if len(args) == 1 {
				gitArgs = []string{"push", remote, args[0]}
			} else {
				gitArgs = []string{"push", remote, "--tags"}
			}

			if flagDryRun {
				output.Infof("[dry-run] git %v", gitArgs)
				return nil
			}
			if _, err := git.Run(cmd.Context(), root, gitArgs...); err != nil {
				return err
			}
			output.Successf("Tags pushed to %s.", remote)
			return nil
		},
	}
}
