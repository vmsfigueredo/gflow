package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newPullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull [remote] [branch]",
		Short: "Pull in all modules",
		Long:  helpPull,
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			remote, branch := flagRemote, ""
			if len(args) >= 1 {
				remote = args[0]
			}
			if len(args) >= 2 {
				branch = args[1]
			}
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, flagPath, flagProject, flagNoRoot)
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

			output.PrintHeader("pull", "", len(mods), output.HeaderFlags{
				Parallel: flagParallel || cfg.Parallel,
				DryRun:   flagDryRun,
			})
			recurse, _ := cmd.Flags().GetBool("recurse-submodules")
			runner := executor.New(cfg, flagParallel, flagFailFast, flagDryRun, flagDebug)
			results := runner.Run(cmd.Context(), mods, func(ctx interface{}, m *module.Module) executor.Result {
				res := git.Pull(m, remote, branch, flagDryRun)
				if res.Status == executor.StatusOK && recurse {
					_ = git.UpdateInit(cmd.Context(), m.Path)
				}
				return res
			})
			return output.Print(results, flagJSON)
		},
	}
	cmd.Flags().BoolP("interactive", "i", false, "pick modules interactively before running")
	cmd.Flags().Bool("recurse-submodules", false, "run submodule update --init --recursive after pull")
	return cmd
}
