package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newCommitCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "commit [git-commit-flags-and-args...]",
		Short:              "Run git commit in all modules (argv passthrough)",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract --interactive from args before forwarding to git.
			// Avoid -i collision with git commit's own -i (include paths) flag.
			interactive := false
			filtered := args[:0]
			for _, a := range args {
				if a == "--interactive" {
					interactive = true
				} else {
					filtered = append(filtered, a)
				}
			}

			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, flagPath, flagProject, flagNoRoot)
			if err != nil {
				return err
			}

			if interactive {
				picked, err := pickInteractive(cmd.Context(), mods)
				if err != nil {
					return err
				}
				if picked != nil {
					mods = picked
				}
			}

			output.PrintHeader("commit", "", len(mods), output.HeaderFlags{
				Parallel: flagParallel || cfg.Parallel,
				DryRun:   flagDryRun,
			})
			runner := executor.New(cfg, flagParallel, flagFailFast, flagDryRun, flagDebug)
			results := runner.Run(cmd.Context(), mods, func(ctx interface{}, m *module.Module) executor.Result {
				return git.Commit(m, filtered, flagDryRun)
			})
			return output.Print(results, flagJSON)
		},
	}
}
