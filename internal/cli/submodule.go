package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newSubmoduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submodule",
		Short: "Manage git submodules (add/remove/move/sync/drift)",
		Long:  helpSubmodule,
	}
	cmd.AddCommand(
		newSubmoduleAddCmd(),
		newSubmoduleRemoveCmd(),
		newSubmoduleMoveCmd(),
		newSubmoduleSyncCmd(),
		newSubmoduleDriftCmd(),
	)
	return cmd
}

func newSubmoduleAddCmd() *cobra.Command {
	var branch string
	cmd := &cobra.Command{
		Use:   "add <url> <path>",
		Short: "Add submodule and register it in config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			url, subPath := args[0], args[1]
			root := resolveRoot()

			output.Infof("Adding submodule %s at %s", url, subPath)
			if flagDryRun {
				output.Infof("[dry-run] git submodule add %s %s", url, subPath)
				return nil
			}
			if err := git.SubmoduleAdd(cmd.Context(), root, url, subPath, branch); err != nil {
				return err
			}
			output.Successf("Submodule added. Update .gflow.conf MODULES array manually or run: gflow config validate")
			return nil
		},
	}
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "branch to track")
	return cmd
}

func newSubmoduleRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <path>",
		Short: "Fully remove a submodule (deinit + git rm + .git/modules cleanup)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			subPath := args[0]

			output.Warnf("This will permanently remove submodule %q from the repository.", subPath)
			if !flagForce {
				ok, err := confirmPrompt(fmt.Sprintf("Remove submodule %q?", subPath))
				if err != nil || !ok {
					return nil
				}
			}
			if flagDryRun {
				output.Infof("[dry-run] git submodule deinit -f %s && git rm -f %s", subPath, subPath)
				return nil
			}
			if err := git.SubmoduleRemove(cmd.Context(), root, subPath); err != nil {
				return err
			}
			output.Successf("Submodule %q removed.", subPath)
			return nil
		},
	}
}

func newSubmoduleMoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "move <old-path> <new-path>",
		Short: "Move a submodule to a new path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			old, newP := args[0], args[1]
			if flagDryRun {
				output.Infof("[dry-run] git mv %s %s && git submodule sync", old, newP)
				return nil
			}
			if err := git.SubmoduleMove(cmd.Context(), root, old, newP); err != nil {
				return err
			}
			output.Successf("Moved submodule %s → %s", old, newP)
			return nil
		},
	}
}

func newSubmoduleSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync submodule URLs from .gitmodules (git submodule sync --recursive)",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			if flagDryRun {
				output.Infof("[dry-run] git submodule sync --recursive")
				return nil
			}
			if err := git.SubmoduleSync(cmd.Context(), root); err != nil {
				return err
			}
			output.Successf("Submodule URLs synced.")
			return nil
		},
	}
}

func newSubmoduleDriftCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drift",
		Short: "List submodules whose HEAD differs from parent-registered pointer",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			_ = cfg

			drifts, err := git.SubmoduleDrift(cmd.Context(), root)
			if err != nil {
				return err
			}
			if len(drifts) == 0 {
				output.Successf("No drift detected — all submodule pointers match HEAD.")
				return nil
			}
			fmt.Printf("%-30s %-10s %-10s %s\n", "SUBMODULE", "REGISTERED", "ACTUAL", "DETACHED")
			for _, d := range drifts {
				reg := short(d.Registered)
				act := short(d.Actual)
				detached := ""
				if d.Detached {
					detached = "yes (detached HEAD)"
				}
				fmt.Printf("%-30s %-10s %-10s %s\n", d.Name, reg, act, detached)
			}
			return nil
		},
	}
}

func resolveRoot() string {
	if flagPath != "" {
		abs, err := filepath.Abs(flagPath)
		if err == nil {
			return abs
		}
	}
	return "."
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}
