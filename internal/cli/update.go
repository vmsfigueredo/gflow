package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/output"
	"github.com/vmsfigueredo/gflow/internal/update"
)

func newUpdateCmd() *cobra.Command {
	var checkOnly, yes, force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for and install gflow updates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			src, exePath, err := update.Detect()
			if err != nil {
				return fmt.Errorf("detect install source: %w", err)
			}

			if src == update.SourceUnknown && !force {
				output.Warnf("cannot determine install source for %s", exePath)
				output.Warnf("use --force to treat as self-managed binary")
				return fmt.Errorf("unknown install source")
			}
			if src == update.SourceUnknown {
				src = update.SourceBinary
			}

			output.Infof("checking for updates...")
			rel, err := update.LatestRelease(ctx, "vmsfigueredo", "gflow")
			if err != nil {
				return fmt.Errorf("fetch latest release: %w", err)
			}

			current := appVersion
			newer, err := update.IsNewer(rel.TagName, current)
			if err != nil {
				return fmt.Errorf("compare versions: %w", err)
			}

			if current == "dev" && !force {
				output.Warnf("running a dev build (%s), use --force to update", exePath)
				return fmt.Errorf("dev build")
			}

			if !newer {
				output.Successf("gflow %s is up to date", current)
				return nil
			}

			output.Infof("current: %s  →  latest: %s  [via %s]", current, rel.TagName, src)

			if rel.Body != "" {
				lines := strings.Split(strings.TrimSpace(rel.Body), "\n")
				limit := 8
				if len(lines) < limit {
					limit = len(lines)
				}
				fmt.Println()
				for _, l := range lines[:limit] {
					fmt.Println(" ", l)
				}
				if len(lines) > limit {
					fmt.Printf("  ... (%d more lines)\n", len(lines)-limit)
				}
				fmt.Println()
			}

			if checkOnly {
				return nil
			}

			if !yes {
				if !isatty.IsTerminal(os.Stdin.Fd()) {
					return fmt.Errorf("non-interactive shell: use --yes to confirm update")
				}
				var confirmed bool
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title(fmt.Sprintf("Update gflow to %s?", rel.TagName)).
							Value(&confirmed),
					),
				)
				if err := form.Run(); err != nil {
					return err
				}
				if !confirmed {
					output.Infof("update cancelled")
					return nil
				}
			}

			if src == update.SourceHomebrew {
				output.Infof("running: brew upgrade gflow")
				return update.ApplyHomebrew(ctx)
			}

			output.Infof("downloading %s...", rel.TagName)
			if err := update.ApplyBinary(ctx, rel, exePath); err != nil {
				return err
			}
			output.Successf("updated to %s", rel.TagName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "only check, don't install")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&force, "force", false, "update even if source detection is ambiguous")
	return cmd
}
