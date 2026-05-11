package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/gh"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newPRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests across modules (requires gh CLI)",
		Long:  helpPR,
	}
	cmd.AddCommand(
		newPRCreateCmd(),
		newPRStatusCmd(),
		newPRMergeCmd(),
	)
	return cmd
}

func newPRCreateCmd() *cobra.Command {
	var (
		base    string
		title   string
		body    string
		draft   bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request in each module that has the current branch published",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !gh.IsAvailable() {
				return fmt.Errorf("gh CLI not found — install: https://cli.github.com")
			}
			root := resolveRoot()
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, root, flagProject, flagNoRoot)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			for _, m := range mods {
				if flagDryRun {
					output.Infof("[%s] [dry-run] gh pr create --base %s", m.Name, base)
					continue
				}
				info, err := gh.PRCreate(ctx, m.Path, title, body, base, draft)
				if err != nil {
					output.Errorf("[%s] %v", m.Name, err)
					continue
				}
				output.Successf("[%s] PR #%d: %s", m.Name, info.Number, info.URL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&base, "base", "develop", "base branch for PR")
	cmd.Flags().StringVar(&title, "title", "", "PR title")
	cmd.Flags().StringVar(&body, "body", "", "PR body")
	cmd.Flags().BoolVar(&draft, "draft", false, "create as draft PR")
	return cmd
}

func newPRStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show PR status for each module",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !gh.IsAvailable() {
				return fmt.Errorf("gh CLI not found — install: https://cli.github.com")
			}
			root := resolveRoot()
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, root, flagProject, flagNoRoot)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "MODULE\tPR#\tSTATE\tMERGEABLE\tURL")
			for _, m := range mods {
				info, err := gh.PRStatus(ctx, m.Path)
				if err != nil {
					fmt.Fprintf(w, "%s\t-\t-\t-\t%v\n", m.Name, err)
					continue
				}
				fmt.Fprintf(w, "%s\t#%d\t%s\t%s\t%s\n", m.Name, info.Number, info.State, info.Mergeable, info.URL)
			}
			return w.Flush()
		},
	}
}

func newPRMergeCmd() *cobra.Command {
	var strategy string
	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge PRs across all modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !gh.IsAvailable() {
				return fmt.Errorf("gh CLI not found — install: https://cli.github.com")
			}
			root := resolveRoot()
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, root, flagProject, flagNoRoot)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			for _, m := range mods {
				if flagDryRun {
					output.Infof("[%s] [dry-run] gh pr merge --%s --auto", m.Name, strategy)
					continue
				}
				if err := gh.PRMerge(ctx, m.Path, strategy); err != nil {
					output.Errorf("[%s] %v", m.Name, err)
					continue
				}
				output.Successf("[%s] merge queued.", m.Name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&strategy, "strategy", "merge", "merge strategy: merge | squash | rebase")
	return cmd
}
