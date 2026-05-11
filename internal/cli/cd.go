package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/registry"
)

func newCdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cd <alias>",
		Short: "Print path for alias (use: cd $(gflow cd <alias>))",
		Long: `Prints the absolute path for a registered project alias.

Add to your shell config:
  gflow-cd() { cd "$(gflow cd "$1")"; }
  alias gcd=gflow-cd
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			r, err := registry.Open()
			if err != nil {
				return err
			}
			e := r.Get(alias)
			if e == nil {
				return fmt.Errorf("alias %q not found — run: gflow projects add %s <path>", alias, alias)
			}
			// Touch last_used silently.
			_ = r.Touch(e.Path)
			fmt.Println(e.Path)
			return nil
		},
	}
}
