package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vmsfigueredo/gflow/internal/output"
)

var appVersion string

// global flags (persistent, available to all subcommands)
var (
	flagPath            string
	flagProject         string
	flagRemote          string
	flagNoRoot          bool
	flagParallel        bool
	flagDryRun          bool
	flagDebug           bool
	flagJSON            bool
	flagNoColor         bool
	flagFailFast        bool
	flagContinueOnError bool
	flagNoAutoCommit    bool
	flagForce           bool
	flagStash           bool
)

func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "gflow",
		Short:         "Git Flow orchestrator for modular repositories",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			output.Init(flagNoColor)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.PersistentFlags().StringVarP(&flagPath, "path", "d", "", "root directory of the repository")
	root.PersistentFlags().StringVarP(&flagProject, "project", "P", "", "resolve alias and use matching modules")
	root.PersistentFlags().StringVarP(&flagRemote, "remote", "R", "origin", "git remote name")
	root.PersistentFlags().BoolVar(&flagNoRoot, "no-root", false, "exclude root module from operations")
	root.PersistentFlags().BoolVar(&flagParallel, "parallel", false, "run operations in parallel across modules")
	root.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "print commands without executing")
	root.PersistentFlags().BoolVar(&flagDebug, "debug", false, "enable debug output")
	root.PersistentFlags().BoolVar(&flagJSON, "json", false, "output as JSON")
	root.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "disable ANSI colors")
	root.PersistentFlags().BoolVar(&flagFailFast, "fail-fast", false, "stop on first module error")
	root.PersistentFlags().BoolVar(&flagContinueOnError, "continue-on-error", false, "continue past errors and aggregate results")
	root.PersistentFlags().BoolVar(&flagNoAutoCommit, "no-auto-commit", false, "skip submodule pointer auto-commit after finish")
	root.PersistentFlags().BoolVar(&flagForce, "force", false, "bypass clean-tree guard")
	root.PersistentFlags().BoolVar(&flagStash, "stash", false, "auto-stash dirty tree before op, pop after")

	root.AddCommand(
		newVersionCmd(version),
		newStatusCmd(),
		newListCmd(),
		newAliasesCmd(),
		newConfigCmd(),
		newInitCmd(),
		newCheckoutCmd(),
		newPullCmd(),
		newPushCmd(),
		newCommitCmd(),
		newFeatureCmd(),
		newHotfixCmd(),
		newBugfixCmd(),
		newReleaseCmd(),
		newDoctorCmd(),
	)
	root.AddCommand(completionCmd(root))
	root.SetHelpFunc(rootHelpFunc)

	return root
}

func rootHelpFunc(cmd *cobra.Command, _ []string) {
	// Groups: gitflow branch ops | git plumbing | helpers/info
	gitflowCmds := []string{"feature", "hotfix", "bugfix", "release", "init"}
	gitCmds := []string{"status", "pull", "push", "commit", "checkout"}
	helperCmds := []string{"list", "aliases", "config", "doctor", "version", "completion"}

	byName := map[string]*cobra.Command{}
	for _, c := range cmd.Commands() {
		byName[c.Name()] = c
	}

	printGroup := func(title string, names []string) {
		output.PrintHelpSection(title, names, byName)
	}

	fmt.Println()
	output.PrintHelpHeader(cmd.Short)
	fmt.Println()
	fmt.Printf("  %s\n\n", output.HelpUsage("gflow <command> [flags]"))

	printGroup("GitFlow Branch Commands", gitflowCmds)
	printGroup("Git Commands", gitCmds)
	printGroup("Tools & Info", helperCmds)

	fmt.Println(output.HelpFlagsSection(cmd))
	fmt.Printf("  Run %s for more information about a command.\n\n",
		output.HelpInlineCode("gflow <command> --help"))
}

func Execute(version string) error {
	appVersion = version
	return newRootCmd(version).Execute()
}
