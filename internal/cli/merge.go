package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
	"github.com/vmsfigueredo/gflow/internal/prompt"
)

func newMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "merge [branch] [git-merge-flags...]",
		Short:              "Run git merge in all modules (argv passthrough)",
		Long:               helpMerge,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Intercept --help/-h before passthrough parsing.
			for _, a := range args {
				if a == "--help" || a == "-h" {
					return cmd.Help()
				}
			}

			interactive := false
			var branch string
			var forwarded []string
			passthroughOnly := false // --abort / --continue: skip branch resolution

			for _, a := range args {
				switch {
				case a == "--interactive":
					interactive = true
				case a == "--abort" || a == "--continue":
					forwarded = append(forwarded, a)
					passthroughOnly = true
				case strings.HasPrefix(a, "-"):
					forwarded = append(forwarded, a)
				case branch == "":
					branch = a
				default:
					forwarded = append(forwarded, a)
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

			if passthroughOnly {
				output.PrintHeader("merge", strings.Join(forwarded, " "), len(mods), output.HeaderFlags{
					Parallel: flagParallel || cfg.Parallel,
					DryRun:   flagDryRun,
				})
				runner := executor.New(cfg, flagParallel, flagFailFast, flagDryRun, flagDebug)
				results := runner.Run(cmd.Context(), mods, func(ctx interface{}, m *module.Module) executor.Result {
					return git.Merge(m, forwarded, flagDryRun)
				})
				return output.Print(results, flagJSON)
			}

			// Interactive module picker when explicitly requested or branch omitted in TTY.
			needModulePick := interactive || (branch == "" && prompt.IsInteractive() && !flagJSON)
			if needModulePick {
				picked, err := pickInteractive(cmd.Context(), mods)
				if err != nil {
					if errors.Is(err, errSilentAbort) {
						return nil
					}
					return err
				}
				if picked != nil {
					mods = picked
				}
			}

			// Interactive branch picker when branch still unknown in TTY.
			if branch == "" && prompt.IsInteractive() && !flagJSON {
				branches := git.BatchLocalBranches(cmd.Context(), mods)
				if len(branches) == 0 {
					return fmt.Errorf("no local branches found")
				}

				// Default cursor: develop > main/master > first.
				def := branches[0]
				for _, candidate := range []string{"develop", "main", "master"} {
					for _, b := range branches {
						if b == candidate {
							def = candidate
							goto foundDef
						}
					}
				}
			foundDef:

				picked, err := prompt.PickBranch("Source branch to merge:", branches, def)
				if err != nil {
					if errors.Is(err, prompt.ErrAborted) {
						fmt.Println("Aborted.")
						return nil
					}
					return err
				}
				branch = picked
			}

			if branch == "" {
				return fmt.Errorf("branch required: pass a positional arg or run in a TTY")
			}

			mergeArgs := append([]string{branch}, forwarded...)
			output.PrintHeader("merge", branch, len(mods), output.HeaderFlags{
				Parallel: flagParallel || cfg.Parallel,
				DryRun:   flagDryRun,
			})
			runner := executor.New(cfg, flagParallel, flagFailFast, flagDryRun, flagDebug)
			results := runner.Run(cmd.Context(), mods, func(ctx interface{}, m *module.Module) executor.Result {
				return git.Merge(m, mergeArgs, flagDryRun)
			})
			return output.Print(results, flagJSON)
		},
	}
}
