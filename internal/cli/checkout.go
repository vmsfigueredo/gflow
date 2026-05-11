package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newCheckoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout <branch>",
		Short: "Checkout branch in all modules",
		Long:  helpCheckout,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := args[0]
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

			output.PrintHeader("checkout", branch, len(mods), output.HeaderFlags{
				Parallel: flagParallel || cfg.Parallel,
				DryRun:   flagDryRun,
			})
			recurse, _ := cmd.Flags().GetBool("recurse-submodules")
			runner := executor.New(cfg, flagParallel, flagFailFast, flagDryRun, flagDebug)
			results := runner.Run(cmd.Context(), mods, func(ctx interface{}, m *module.Module) executor.Result {
				res := git.Checkout(m, branch, flagDryRun)
				if res.Status == executor.StatusOK && recurse {
					_ = git.UpdateInit(cmd.Context(), m.Path)
				}
				return res
			})
			return output.Print(results, flagJSON)
		},
	}
	cmd.Flags().BoolP("interactive", "i", false, "pick modules interactively before running")
	cmd.Flags().Bool("recurse-submodules", false, "run submodule update --init --recursive after checkout")
	return cmd
}
