package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/gitflow"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize git-flow in all modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, flagPath, flagProject, flagNoRoot)
			if err != nil {
				return err
			}
			output.PrintHeader("init", "", len(mods), output.HeaderFlags{
				Parallel: flagParallel || cfg.Parallel,
				DryRun:   flagDryRun,
			})
			results, err := gitflow.Init(cmd.Context(), cfg, mods, flagDryRun, flagDebug)
			if err != nil {
				return err
			}
			return output.Print(results, flagJSON)
		},
	}
}
