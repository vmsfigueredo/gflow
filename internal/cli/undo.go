package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/journal"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newUndoCmd() *cobra.Command {
	var targetModule string
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Revert the last recorded gflow operation (resets modules to pre-op SHAs)",
		Long:  helpUndo,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			j, err := journal.Open(root)
			if err != nil {
				return err
			}
			last, err := j.Last()
			if err != nil {
				return err
			}
			if last == nil {
				return fmt.Errorf("no operations recorded — nothing to undo")
			}

			output.Warnf("Undoing: %s (at %s)", last.Op, last.Timestamp.Format("2006-01-02 15:04:05"))
			output.Warnf("This will reset HEAD for each affected module to its pre-operation SHA.")

			if !flagForce {
				ok, err := confirmPrompt(fmt.Sprintf("Undo %q?", last.Op))
				if err != nil || !ok {
					output.Infof("Aborted.")
					return nil
				}
			}

			ctx := cmd.Context()
			for mod, sha := range last.RefsBefore {
				if targetModule != "" && mod != targetModule {
					continue
				}
				modPath := resolveModulePath(root, mod)
				output.Infof("[%s] reset to %s", mod, sha[:8])
				if flagDryRun {
					continue
				}
				if _, err := git.Run(ctx, modPath, "reset", "--hard", sha); err != nil {
					output.Errorf("[%s] reset failed: %v", mod, err)
				}
			}

			output.Successf("Undo complete.")
			return nil
		},
	}
	cmd.Flags().StringVar(&targetModule, "module", "", "undo only for this module name")
	return cmd
}

func resolveModulePath(root, modName string) string {
	// If root itself matches, return root.
	// Otherwise assume submodule at root/<modName>.
	import_path := root + "/" + modName
	// Check if root is the module (single-repo case).
	if modName == "." || modName == "" {
		return root
	}
	return import_path
}
