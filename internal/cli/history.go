package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/journal"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newHistoryCmd() *cobra.Command {
	var n int
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show gflow operation history",
		Long:  helpHistory,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := resolveRoot()
			j, err := journal.Open(root)
			if err != nil {
				return err
			}
			entries, err := j.ReadAll()
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				output.Infof("No operations recorded yet.")
				return nil
			}
			// Show last n.
			if n > 0 && n < len(entries) {
				entries = entries[len(entries)-n:]
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTIME\tOP\tMODULES\tSTATUS")
			for _, e := range entries {
				status := summaryStatus(e)
				mods := summarizeList(e.Modules, 3)
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					e.ID[:8],
					e.Timestamp.Format("2006-01-02 15:04"),
					e.Op,
					mods,
					status,
				)
			}
			return w.Flush()
		},
	}
	cmd.Flags().IntVarP(&n, "count", "n", 20, "number of entries to show")
	return cmd
}

func summaryStatus(e journal.Entry) string {
	ok, errCount := 0, 0
	for _, r := range e.Results {
		if r.Status == "error" {
			errCount++
		} else if r.Status == "ok" {
			ok++
		}
	}
	if errCount > 0 {
		return fmt.Sprintf("%d ok / %d err", ok, errCount)
	}
	return fmt.Sprintf("%d ok", ok)
}

func summarizeList(items []string, max int) string {
	if len(items) <= max {
		s := ""
		for i, item := range items {
			if i > 0 {
				s += ","
			}
			s += item
		}
		return s
	}
	s := ""
	for i := 0; i < max; i++ {
		if i > 0 {
			s += ","
		}
		s += items[i]
	}
	return fmt.Sprintf("%s+%d", s, len(items)-max)
}
