package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/flow"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
	"github.com/vmsfigueredo/gflow/internal/prompt"
)

func newFlowTypeCmd(branchType string) *cobra.Command {
	long := ""
	switch branchType {
	case "feature":
		long = helpFeature
	case "hotfix":
		long = helpHotfix
	case "bugfix":
		long = helpBugfix
	case "release":
		long = helpRelease
	}
	cmd := &cobra.Command{
		Use:   branchType,
		Short: "Manage " + branchType + " branches across modules",
		Long:  long,
	}

	cmd.AddCommand(
		newFlowOpCmd(branchType, "start", "[name]", cobra.MaximumNArgs(1)),
		newFlowOpCmd(branchType, "finish", "[name]", cobra.MaximumNArgs(1)),
		newFlowOpCmd(branchType, "publish", "[name]", cobra.MaximumNArgs(1)),
		newFlowOpCmd(branchType, "track", "[name]", cobra.MaximumNArgs(1)),
		newFlowOpCmd(branchType, "delete", "[name]", cobra.MaximumNArgs(1)),
	)

	// Per-type flags cascade to op subcommands.
	cmd.PersistentFlags().BoolP("interactive", "i", false, "force interactive module picker")
	cmd.PersistentFlags().String("names", "", "per-module names: api=v1.2.3,web=v1.4.0")

	return cmd
}

func newFlowOpCmd(branchType, op, argUsage string, argsValidator cobra.PositionalArgs) *cobra.Command {
	return &cobra.Command{
		Use:   op + " " + argUsage,
		Short: op + " a " + branchType + " branch across all modules",
		Args:  argsValidator,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve flags (inherited from parent).
			iFlag, _ := cmd.Flags().GetBool("interactive")
			namesStr, _ := cmd.Flags().GetString("names")

			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, flagPath, flagProject, flagNoRoot)
			if err != nil {
				return err
			}

			namesMap := parseNamesFlag(namesStr)

			// Positional arg, if supplied.
			positional := ""
			if len(args) > 0 {
				positional = args[0]
			}

			// Decide whether to enter interactive mode.
			needInteractive := (positional == "" && len(namesMap) == 0 || iFlag) &&
				prompt.IsInteractive() && !flagJSON

			opts := flow.Options{
				BranchType:   branchType,
				Op:           op,
				Remote:       flagRemote,
				Parallel:     flagParallel,
				FailFast:     flagFailFast,
				DryRun:       flagDryRun,
				Debug:        flagDebug,
				Force:        flagForce,
				Stash:        flagStash,
				NoAutoCommit: flagNoAutoCommit,
			}

			if !needInteractive {
				// Non-interactive path.
				if positional == "" && len(namesMap) == 0 {
					return errors.New(
						"name required: pass a positional arg, --names api=foo,web=bar, or run in a TTY",
					)
				}
				opts.Name = positional
				opts.NamesByModule = namesMap
				output.PrintHeader(branchType+" "+op, positional, len(mods), output.HeaderFlags{
					Parallel: flagParallel || cfg.Parallel,
					DryRun:   flagDryRun,
				})
				results, err := flow.Run(cmd.Context(), cfg, mods, opts)
				if err != nil {
					return err
				}
				return output.Print(results, flagJSON)
			}

			// ── Interactive path ──────────────────────────────────────────────

			ctx := cmd.Context()

			// Fetch current branch for every module in parallel (picker hints).
			hints := git.BatchCurrentBranches(ctx, mods)

			// Step 1: pick modules (all pre-selected).
			picked, err := prompt.PickModules(mods, hints)
			if err != nil {
				if errors.Is(err, prompt.ErrAborted) {
					return nil
				}
				return err
			}
			if len(picked) == 0 {
				fmt.Println("No modules selected. Aborted.")
				return nil
			}

			// Step 2: resolve names per module.
			var namesByMod map[string]string

			prefix := branchTypePrefix(cfg, branchType)

			switch op {
			case "start":
				// For start, prompt a name for each picked module.
				defaults := make(map[string]string, len(picked))
				for _, m := range picked {
					if positional != "" {
						defaults[m.Name] = positional
					} else if namesMap != nil {
						defaults[m.Name] = namesMap[m.Name]
					}
				}
				namesByMod, err = prompt.AskPerModuleName(branchType, picked, defaults)
				if err != nil {
					if errors.Is(err, prompt.ErrAborted) {
						return nil
					}
					return err
				}

			default:
				// finish / publish / track / delete:
				// Auto-detect from current branch per module; prompt only the unresolved.
				namesByMod = make(map[string]string, len(picked))
				var unresolved []*module.Module

				for _, m := range picked {
					if n, ok, _ := git.DetectCurrentFlowBranch(ctx, m.Path, prefix); ok {
						namesByMod[m.Name] = n
					} else {
						unresolved = append(unresolved, m)
					}
				}

				if len(unresolved) > 0 {
					defaults := make(map[string]string)
					if namesMap != nil {
						for _, m := range unresolved {
							defaults[m.Name] = namesMap[m.Name]
						}
					}
					extra, err := prompt.AskPerModuleName(branchType, unresolved, defaults)
					if err != nil {
						if errors.Is(err, prompt.ErrAborted) {
							return nil
						}
						return err
					}
					for k, v := range extra {
						namesByMod[k] = v
					}
				}
			}

			if len(namesByMod) == 0 {
				fmt.Println("No modules with a name. Aborted.")
				return nil
			}

			// Step 3: confirm.
			summary := buildConfirmSummary(op, branchType, namesByMod, flagDryRun, flagParallel || cfg.Parallel)
			ok, err := prompt.ConfirmSummary(summary)
			if err != nil || !ok {
				if err != nil && !errors.Is(err, prompt.ErrAborted) {
					return err
				}
				fmt.Println("Aborted.")
				return nil
			}

			// Step 4: run.
			opts.NamesByModule = namesByMod
			output.PrintHeader(branchType+" "+op, "", len(namesByMod), output.HeaderFlags{
				Parallel: flagParallel || cfg.Parallel,
				DryRun:   flagDryRun,
			})
			results, err := flow.Run(ctx, cfg, mods, opts)
			if err != nil {
				return err
			}
			return output.Print(results, flagJSON)
		},
	}
}

// parseNamesFlag parses "api=v1.2.3,web=v1.4.0" into a map.
// Returns nil on empty input.
func parseNamesFlag(s string) map[string]string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	m := map[string]string{}
	for _, pair := range strings.Split(s, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// branchTypePrefix returns the configured git-flow branch prefix for a type.
func branchTypePrefix(cfg *config.Config, branchType string) string {
	switch branchType {
	case "feature":
		return cfg.Gitflow.FeaturePrefix
	case "hotfix":
		return cfg.Gitflow.HotfixPrefix
	case "bugfix":
		return cfg.Gitflow.BugfixPrefix
	case "release":
		return cfg.Gitflow.ReleasePrefix
	default:
		return branchType + "/"
	}
}

// buildConfirmSummary formats the confirmation line shown before running.
func buildConfirmSummary(op, branchType string, names map[string]string, dryRun, parallel bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s on %d module(s):\n", branchType, op, len(names)))
	for mod, name := range names {
		b.WriteString(fmt.Sprintf("  %-20s → %s\n", mod, name))
	}
	chips := []string{}
	if parallel {
		chips = append(chips, "parallel")
	}
	if dryRun {
		chips = append(chips, "dry-run")
	}
	if len(chips) > 0 {
		b.WriteString("  [" + strings.Join(chips, ", ") + "]")
	}
	return strings.TrimRight(b.String(), "\n")
}

func newFeatureCmd() *cobra.Command {
	cmd := newFlowTypeCmd("feature")
	cmd.AddCommand(newFeatureUpdateCmd())
	return cmd
}
func newHotfixCmd() *cobra.Command  { return newFlowTypeCmd("hotfix") }
func newBugfixCmd() *cobra.Command  { return newFlowTypeCmd("bugfix") }
func newReleaseCmd() *cobra.Command { return newFlowTypeCmd("release") }

func newFeatureUpdateCmd() *cobra.Command {
	var useRebase bool
	var useMerge bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Integrate develop into the current feature branch (fetch + merge or rebase)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, flagPath, flagProject, flagNoRoot)
			if err != nil {
				return err
			}

			// Determine strategy: flag > config > default merge.
			strategy := cfg.Gitflow.FeatureStrategy
			if useRebase {
				strategy = "rebase"
			} else if useMerge {
				strategy = "merge"
			}
			if strategy == "" {
				strategy = "merge"
			}

			ctx := cmd.Context()
			output.PrintHeader("feature update", "("+strategy+")", len(mods), output.HeaderFlags{
				Parallel: flagParallel || cfg.Parallel,
				DryRun:   flagDryRun,
			})

			runner := executor.New(cfg, flagParallel, flagFailFast, flagDryRun, flagDebug)
			results := runner.Run(ctx, mods, func(_ interface{}, m *module.Module) executor.Result {
				return featureUpdate(ctx, m, cfg.Remote, strategy, flagDryRun)
			})
			return output.Print(results, flagJSON)
		},
	}
	cmd.Flags().BoolVar(&useRebase, "rebase", false, "rebase feature onto develop (overrides config)")
	cmd.Flags().BoolVar(&useMerge, "merge", false, "merge develop into feature (overrides config)")
	return cmd
}

// featureUpdate fetches remote develop and integrates it into the current branch.
func featureUpdate(ctx context.Context, m *module.Module, remote, strategy string, dryRun bool) executor.Result {
	if dryRun {
		return executor.Result{
			Module: m, Action: "feature update",
			Status: executor.StatusDryRun,
			Output: fmt.Sprintf("git fetch %s develop && git %s develop", remote, strategy),
		}
	}

	// Fetch develop from remote.
	if _, err := git.Run(ctx, m.Path, "fetch", remote, "develop"); err != nil {
		return executor.Result{Module: m, Action: "feature update", Status: executor.StatusError, Err: err}
	}

	switch strategy {
	case "rebase":
		if _, err := git.Run(ctx, m.Path, "rebase", remote+"/develop"); err != nil {
			return executor.Result{Module: m, Action: "feature update", Status: executor.StatusError, Err: err}
		}
		// Force-push with lease after rebase.
		branch, _ := git.CurrentBranch(ctx, m.Path)
		if _, err := git.Run(ctx, m.Path, "push", "--force-with-lease", remote, branch); err != nil {
			return executor.Result{Module: m, Action: "feature update", Status: executor.StatusError, Err: err}
		}
	default:
		if _, err := git.Run(ctx, m.Path, "merge", remote+"/develop"); err != nil {
			return executor.Result{Module: m, Action: "feature update", Status: executor.StatusError, Err: err}
		}
	}

	return executor.Result{Module: m, Action: "feature update", Status: executor.StatusOK}
}
