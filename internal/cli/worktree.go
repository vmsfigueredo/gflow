package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newWorktreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worktree",
		Short: "Manage git worktrees across modules (requires git ≥ 2.38)",
		Long:  helpWorktree,
	}
	cmd.AddCommand(
		newWorktreeAddCmd(),
		newWorktreeListCmd(),
		newWorktreeRemoveCmd(),
		newWorktreeSwitchCmd(),
	)
	return cmd
}

func newWorktreeAddCmd() *cobra.Command {
	var targetPath string
	cmd := &cobra.Command{
		Use:   "add <branch> [--path <dir>]",
		Short: "Create a worktree for branch across all modules",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := args[0]
			root := resolveRoot()
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, root, flagProject, flagNoRoot)
			if err != nil {
				return err
			}

			for _, m := range mods {
				dest := targetPath
				if dest == "" {
					dest = m.Path + "-" + branch
				}
				output.Infof("[%s] worktree add %s → %s", m.Name, branch, dest)
				if flagDryRun {
					continue
				}
				if err := git.WorktreeAdd(cmd.Context(), m.Path, dest, branch); err != nil {
					output.Errorf("[%s] %v", m.Name, err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&targetPath, "path", "", "target directory (default: <module>-<branch>)")
	return cmd
}

func newWorktreeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all worktrees across modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, root, flagProject, flagNoRoot)
			if err != nil {
				return err
			}

			for _, m := range mods {
				entries, err := git.WorktreeList(cmd.Context(), m.Path)
				if err != nil {
					output.Warnf("[%s] %v", m.Name, err)
					continue
				}
				for _, e := range entries {
					branch := e.Branch
					if branch == "" {
						branch = "(detached)"
					}
					fmt.Printf("%-20s %-30s %s\n", m.Name, branch, e.Path)
				}
			}
			return nil
		},
	}
}

func newWorktreeRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <path>",
		Short: "Remove a worktree and prune stale entries",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			targetPath := args[0]
			if flagDryRun {
				output.Infof("[dry-run] git worktree remove --force %s && git worktree prune", targetPath)
				return nil
			}
			if err := git.WorktreeRemove(cmd.Context(), root, targetPath); err != nil {
				return err
			}
			output.Successf("Worktree %s removed.", targetPath)
			return nil
		},
	}
}

func newWorktreeSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <branch>",
		Short: "Print worktree path for branch (use: cd $(gflow worktree switch <branch>))",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := args[0]
			root := resolveRoot()
			entries, err := git.WorktreeList(cmd.Context(), root)
			if err != nil {
				return err
			}
			for _, e := range entries {
				if e.Branch == branch {
					fmt.Println(e.Path)
					return nil
				}
			}
			return fmt.Errorf("no worktree found for branch %q", branch)
		},
	}
}
