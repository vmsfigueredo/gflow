package cli

import (
	"fmt"

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
			fmt.Println(res.Stdout)
			return nil
		},
	}
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
