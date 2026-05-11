package cli

import (
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/doctor"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check environment and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flagPath)
			if err != nil {
				return err
			}
			return doctor.Run(cmd.Context(), cfg, flagPath, appVersion, flagJSON)
		},
	}
}
