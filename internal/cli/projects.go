package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/output"
	"github.com/vmsfigueredo/gflow/internal/registry"
)

func newProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage global project registry (~/.config/gflow/projects.yaml)",
		Long:  helpProjects,
	}
	cmd.AddCommand(
		newProjectsAddCmd(),
		newProjectsListCmd(),
		newProjectsRemoveCmd(),
		newProjectsRecentCmd(),
	)
	return cmd
}

func newProjectsAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <alias> [path]",
		Short: "Register a project (defaults to current directory)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			var absPath string
			if len(args) == 2 {
				p, err := filepath.Abs(args[1])
				if err != nil {
					return err
				}
				absPath = p
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				absPath = cwd
			}

			r, err := registry.Open()
			if err != nil {
				return err
			}
			if err := r.Add(alias, absPath); err != nil {
				return err
			}
			output.Successf("Registered %q → %s", alias, absPath)
			output.Infof("Tip: add to your shell: gflow-cd() { cd \"$(gflow cd $1)\"; }")
			return nil
		},
	}
}

func newProjectsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Open()
			if err != nil {
				return err
			}
			if len(r.Entries) == 0 {
				fmt.Println("No projects registered. Run: gflow projects add <alias> [path]")
				return nil
			}

			ctx := cmd.Context()
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ALIAS\tPATH\tBRANCH\tDIRTY\tLAST USED")
			for _, e := range r.Entries {
				branch, _ := git.CurrentBranch(ctx, e.Path)
				dirty, _ := git.IsDirty(ctx, e.Path)
				dirtyStr := ""
				if dirty {
					dirtyStr = "*"
				}
				lastUsed := ""
				if !e.LastUsed.IsZero() {
					lastUsed = e.LastUsed.Format(time.RFC3339)[:10]
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", e.Alias, e.Path, branch, dirtyStr, lastUsed)
			}
			return w.Flush()
		},
	}
}

func newProjectsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <alias>",
		Short: "Remove a registered project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Open()
			if err != nil {
				return err
			}
			if err := r.Remove(args[0]); err != nil {
				return err
			}
			output.Successf("Removed %q from registry.", args[0])
			return nil
		},
	}
}

func newProjectsRecentCmd() *cobra.Command {
	var n int
	cmd := &cobra.Command{
		Use:   "recent",
		Short: "Show recently used projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := registry.Open()
			if err != nil {
				return err
			}
			entries := r.Recent(n)
			if len(entries) == 0 {
				fmt.Println("No recent projects.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ALIAS\tPATH\tLAST USED")
			for _, e := range entries {
				fmt.Fprintf(w, "%s\t%s\t%s\n", e.Alias, e.Path, e.LastUsed.Format(time.RFC3339)[:10])
			}
			return w.Flush()
		},
	}
	cmd.Flags().IntVarP(&n, "count", "n", 10, "number of entries to show")
	return cmd
}
