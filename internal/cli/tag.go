package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

const tagListLimit = 5

type tagRow struct {
	name string
	tags []string
}

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage git tags (list, push)",
	}
	cmd.AddCommand(
		newTagListCmd(),
		newTagPushCmd(),
	)
	return cmd
}

func newTagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tags across all modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, flagPath, flagProject, flagNoRoot)
			if err != nil {
				return err
			}

			rows := make([]tagRow, 0, len(mods))
			for _, m := range mods {
				res, err := git.Run(cmd.Context(), m.Path, "tag", "--sort=-version:refname")
				if err != nil {
					output.Warnf("%s: %v", m.Name, err)
					rows = append(rows, tagRow{name: m.Name, tags: nil})
					continue
				}
				all := strings.Fields(strings.TrimSpace(res.Stdout))
				if len(all) > tagListLimit {
					all = all[:tagListLimit]
				}
				rows = append(rows, tagRow{name: m.Name, tags: all})
			}

			printTagTable(rows)
			return nil
		},
	}
}

func printTagTable(rows []tagRow) {
	// compute column widths: module name + up to tagListLimit tag columns
	nameW := len("MODULE")
	for _, r := range rows {
		if len(r.name) > nameW {
			nameW = len(r.name)
		}
	}

	colW := make([]int, tagListLimit)
	headers := []string{"LATEST", "TAG 2", "TAG 3", "TAG 4", "TAG 5"}
	for i, h := range headers {
		colW[i] = len(h)
	}
	for _, r := range rows {
		for i, t := range r.tags {
			w := len(t)
			if i == 0 {
				w += len(" (latest)")
			}
			if w > colW[i] {
				colW[i] = w
			}
		}
	}

	sep := "  "

	// header
	fmt.Println()
	line := output.HelpInlineCode(fmt.Sprintf("%-*s", nameW, "MODULE"))
	for i, h := range headers {
		line += sep + output.HelpUsage(fmt.Sprintf("%-*s", colW[i], h))
	}
	fmt.Println("  " + line)
	fmt.Println("  " + strings.Repeat("─", nameW+2+sumInts(colW)+len(sep)*tagListLimit))

	// rows
	for _, r := range rows {
		if r.tags == nil {
			fmt.Printf("  %-*s  %s\n", nameW, r.name, output.HelpUsage("—"))
			continue
		}
		cells := make([]string, tagListLimit)
		for i := 0; i < tagListLimit; i++ {
			if i < len(r.tags) {
				t := r.tags[i]
				if i == 0 {
					cells[i] = fmt.Sprintf("%-*s", colW[i], t+" (latest)")
				} else {
					cells[i] = fmt.Sprintf("%-*s", colW[i], t)
				}
			} else {
				cells[i] = fmt.Sprintf("%-*s", colW[i], "—")
			}
		}
		fmt.Printf("  %-*s%s%s\n", nameW, r.name, sep, strings.Join(cells, sep))
	}
	fmt.Println()
}

func sumInts(s []int) int {
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

func newTagPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push [tag]",
		Short: "Push tags to remote (all tags if no arg)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			remote := flagRemote

			var gitArgs []string
			if len(args) == 1 {
				gitArgs = []string{"push", remote, args[0]}
			} else {
				gitArgs = []string{"push", remote, "--tags"}
			}

			if flagDryRun {
				output.Infof("[dry-run] git %v", gitArgs)
				return nil
			}
			if _, err := git.Run(cmd.Context(), root, gitArgs...); err != nil {
				return err
			}
			output.Successf("Tags pushed to %s.", remote)
			return nil
		},
	}
}
