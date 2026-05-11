package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/output"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			mods, err := module.Resolve(cfg, flagPath, flagProject, flagNoRoot)
			if err != nil {
				return err
			}
			return output.PrintList(mods, flagJSON)
		},
	}
}
