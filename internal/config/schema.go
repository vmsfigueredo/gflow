package config

// Config is the unified configuration for gflow.
type Config struct {
	Modules  []string            `yaml:"modules"`
	Aliases  map[string][]string `yaml:"aliases"`
	Parallel bool                `yaml:"parallel"`
	Workers  int                 `yaml:"parallel_workers"`
	Debug    bool                `yaml:"debug"`
	Remote   string              `yaml:"remote"`
	NoRoot   bool                `yaml:"no_root"`

	Gitflow GitflowConfig `yaml:"gitflow"`
	Submod  SubmodConfig  `yaml:"submodule"`

	BranchExistsPolicy string `yaml:"branch_exists_policy"` // skip | checkout | error
}

type GitflowConfig struct {
	FinishOptions    string `yaml:"finish_options"`
	FeaturePrefix    string `yaml:"feature_prefix"`
	HotfixPrefix     string `yaml:"hotfix_prefix"`
	BugfixPrefix     string `yaml:"bugfix_prefix"`
	ReleasePrefix    string `yaml:"release_prefix"`
	SupportPrefix    string `yaml:"support_prefix"`
	VersionTagPrefix string `yaml:"version_tag_prefix"`
	// FeatureStrategy controls how develop is integrated into feature branches.
	// Valid values: "merge" (default) | "rebase"
	FeatureStrategy string `yaml:"feature_strategy"`
}

type SubmodConfig struct {
	AutoCommit    bool   `yaml:"auto_commit"`
	CommitMessage string `yaml:"commit_message"`
}

func defaults() Config {
	return Config{
		Remote:  "origin",
		Workers: 0, // 0 = runtime.NumCPU()
		Gitflow: GitflowConfig{
			FeaturePrefix:    "feature/",
			HotfixPrefix:     "hotfix/",
			BugfixPrefix:     "bugfix/",
			ReleasePrefix:    "release/",
			SupportPrefix:    "support/",
			VersionTagPrefix: "v",
			FeatureStrategy:  "merge",
		},
		Submod: SubmodConfig{
			AutoCommit:    true,
			CommitMessage: "chore: bump {module} to {sha}",
		},
		BranchExistsPolicy: "checkout",
	}
}
