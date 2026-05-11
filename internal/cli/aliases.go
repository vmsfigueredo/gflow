package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newAliasesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "aliases",
		Short: "Print configured module aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			return output.PrintAliases(cfg.Aliases, flagJSON)
		},
	}
}
