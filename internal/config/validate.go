package config

import (
	"errors"
	"fmt"
)

func validate(cfg *Config) error {
	valid := map[string]bool{"skip": true, "checkout": true, "error": true}
	if cfg.BranchExistsPolicy != "" && !valid[cfg.BranchExistsPolicy] {
		return fmt.Errorf("branch_exists_policy %q invalid: must be skip|checkout|error", cfg.BranchExistsPolicy)
	}
	if cfg.Workers < 0 {
		return errors.New("parallel_workers must be >= 0")
	}
	return nil
}
