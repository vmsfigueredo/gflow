package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
)

func newConfigCmd() *cobra.Command {
	var emitYAML, emitBash bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Generate or manage gflow configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			if emitBash {
				return config.EmitBash(cfg, flagPath)
			}
			return config.EmitYAML(cfg, flagPath)
		},
	}
	cmd.Flags().BoolVar(&emitYAML, "yaml", false, "emit gflow.yaml format")
	cmd.Flags().BoolVar(&emitBash, "bash", false, "emit .gflow.conf bash format")

	cmd.AddCommand(newConfigMigrateCmd(), newConfigValidateCmd())
	return cmd
}

func newConfigMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Convert .gflow.conf to gflow.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.Migrate(flagPath)
		},
	}
}

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate active configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			return config.Validate(cfg)
		},
	}
}
