package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/flow"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newFlowTypeCmd(branchType string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   branchType,
		Short: "Manage " + branchType + " branches across modules",
	}

	cmd.AddCommand(
		newFlowOpCmd(branchType, "start", "<name>", cobra.ExactArgs(1)),
		newFlowOpCmd(branchType, "finish", "<name>", cobra.ExactArgs(1)),
		newFlowOpCmd(branchType, "publish", "<name>", cobra.ExactArgs(1)),
		newFlowOpCmd(branchType, "track", "<name>", cobra.ExactArgs(1)),
		newFlowOpCmd(branchType, "delete", "<name>", cobra.ExactArgs(1)),
	)
	return cmd
}

func newFlowOpCmd(branchType, op, argUsage string, argsValidator cobra.PositionalArgs) *cobra.Command {
	return &cobra.Command{
		Use:   op + " " + argUsage,
		Short: op + " a " + branchType + " branch across all modules",
		Args:  argsValidator,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, flagPath, flagProject, flagNoRoot)
			if err != nil {
				return err
			}
			opts := flow.Options{
				BranchType:   branchType,
				Op:           op,
				Name:         name,
				Remote:       flagRemote,
				Parallel:     flagParallel,
				FailFast:     flagFailFast,
				DryRun:       flagDryRun,
				Debug:        flagDebug,
				Force:        flagForce,
				Stash:        flagStash,
				NoAutoCommit: flagNoAutoCommit,
			}
			output.PrintHeader(branchType+" "+op, name, len(mods), output.HeaderFlags{
				Parallel: flagParallel || cfg.Parallel,
				DryRun:   flagDryRun,
			})
			results, err := flow.Run(cmd.Context(), cfg, mods, opts)
			if err != nil {
				return err
			}
			return output.Print(results, flagJSON)
		},
	}
}

func newFeatureCmd() *cobra.Command { return newFlowTypeCmd("feature") }
func newHotfixCmd() *cobra.Command  { return newFlowTypeCmd("hotfix") }
func newBugfixCmd() *cobra.Command  { return newFlowTypeCmd("bugfix") }
func newReleaseCmd() *cobra.Command { return newFlowTypeCmd("release") }
