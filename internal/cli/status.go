package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show git status for all modules",
		Long:  helpStatus,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, flagPath, flagProject, flagNoRoot, flagExcept)
			if err != nil {
				return err
			}

			iFlag, _ := cmd.Flags().GetBool("interactive")
			if iFlag {
				picked, err := pickInteractive(cmd.Context(), mods)
				if err != nil {
					return err
				}
				if picked != nil {
					mods = picked
				}
			}

			output.PrintHeader("status", "", len(mods), output.HeaderFlags{
				Parallel: flagParallel || cfg.Parallel,
				DryRun:   flagDryRun,
			})
			runner := executor.New(cfg, flagParallel, flagFailFast, flagDryRun, flagDebug)
			results := runner.Run(cmd.Context(), mods, func(ctx interface{}, m *module.Module) executor.Result {
				return git.Status(m, flagDryRun)
			})
			return output.Print(results, flagJSON)
		},
	}
	cmd.Flags().BoolP("interactive", "i", false, "pick modules interactively before running")
	return cmd
}
