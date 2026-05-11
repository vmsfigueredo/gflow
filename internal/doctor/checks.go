package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vmsfigueredo/gflow/internal/config"
	"github.com/vmsfigueredo/gflow/internal/gitflow"
	"github.com/vmsfigueredo/gflow/internal/output"
)

type check struct {
	label  string
	run    func() (bool, string)
}

// Run executes all doctor checks and reports results.
// Uses appVersion (injected from main) — no hardcoded version strings.
func Run(ctx context.Context, cfg *config.Config, root, appVersion string, asJSON bool) error {
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	checks := []check{
		{
			label: "git installed",
			run: func() (bool, string) {
				_, err := exec.LookPath("git")
				return err == nil, "git not found in PATH"
			},
		},
		{
			label: "git-flow installed",
			run: func() (bool, string) {
				_, err := exec.CommandContext(ctx, "git", "flow", "version").Output()
				return err == nil, "git flow not found — install via `brew install git-flow-avh` (macOS) or package manager"
			},
		},
		{
			label: "git-flow variant detected",
			run: func() (bool, string) {
				v, err := gitflow.DetectVariant(ctx)
				if err != nil {
					return false, err.Error()
				}
				return v != gitflow.VariantUnknown, fmt.Sprintf("variant: %s", v)
			},
		},
		{
			label: "config file found",
			run: func() (bool, string) {
				yaml := filepath.Join(root, "gflow.yaml")
				bash := filepath.Join(root, ".gflow.conf")
				switch {
				case fileExists(yaml):
					return true, "gflow.yaml"
				case fileExists(bash):
					return true, ".gflow.conf (legacy)"
				default:
					return false, "no gflow.yaml or .gflow.conf — using defaults"
				}
			},
		},
		{
			label: "config valid",
			run: func() (bool, string) {
				return config.Validate(cfg) == nil, "run `gflow config validate` for details"
			},
		},
		{
			label: ".gitmodules present",
			run: func() (bool, string) {
				ok := fileExists(filepath.Join(root, ".gitmodules"))
				if !ok {
					return false, "not a submodule repo (ok if using MODULES list)"
				}
				return true, ""
			},
		},
		{
			label: fmt.Sprintf("gflow version %s", appVersion),
			run:   func() (bool, string) { return true, "" },
		},
	}

	failed, passed := 0, 0
	for _, c := range checks {
		ok, detail := c.run()
		if ok {
			passed++
		} else {
			failed++
		}
		output.PrintDoctorCheck(c.label, ok, detail)
	}

	output.DoctorSummary(passed, failed)

	if failed > 0 {
		return fmt.Errorf("%d check(s) failed", failed)
	}
	return nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
