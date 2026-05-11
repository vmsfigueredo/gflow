package cli

// Rich Long descriptions for each command group.
// Injected via cmd.Long in each command constructor.

const helpFeature = `Manage feature branches across all modules.

A feature branch follows git-flow conventions (feature/<name>) and is
always branched from develop. gflow orchestrates the operation across
every module (root + submodules) in one shot.

Subcommands:
  start    Create feature/<name> from develop in all modules
  finish   Merge feature into develop, delete branch, bump submodule pointer
  publish  Push feature to remote so collaborators can track it
  track    Set up local tracking for a remote feature branch
  delete   Delete feature branch locally (and optionally remotely)
  update   Fetch + integrate develop into current feature (merge or rebase)

Examples:
  gflow feature start my-feature
  gflow feature start my-feature --parallel
  gflow feature finish my-feature
  gflow feature finish my-feature --no-auto-commit
  gflow feature update --rebase
  gflow feature update --merge
  gflow feature publish my-feature
  gflow feature delete my-feature

  # Per-module names (e.g. api gets v1, web gets v2):
  gflow feature start --names api=v1,web=v2

  # Interactive picker (select which modules to operate on):
  gflow feature start -i my-feature

Flags:
  --parallel           Run across modules concurrently
  --fail-fast          Stop on first module error
  --dry-run            Print commands without executing
  --force              Skip clean-tree guard
  --stash              Auto-stash dirty working tree, pop after
  --no-auto-commit     Skip submodule pointer commit in parent after finish
  --names k=v,...      Per-module branch names
  -i, --interactive    Pick modules via TUI before running

Config (gflow.yaml):
  gitflow:
    feature_prefix: "feature/"
    feature_strategy: "merge"   # or "rebase" (used by feature update)
`

const helpHotfix = `Manage hotfix branches across all modules.

Hotfix branches are created from main/master to patch production bugs.
gflow validates semver on start and coordinates finish across all modules.

Subcommands:
  start    Create hotfix/<version> from main in all modules
  finish   Merge into main + develop, create tag, bump submodule pointer
  publish  Push hotfix to remote
  track    Track remote hotfix locally
  delete   Delete hotfix branch

Examples:
  gflow hotfix start 1.2.1
  gflow hotfix finish 1.2.1
  gflow hotfix publish 1.2.1

Guards:
  - SemverGuard: version must be valid semver (e.g. 1.2.1, not "next")
  - OnExpectedBranchGuard: must be on main/master before start
  - RemoteSyncGuard: main must be in sync with remote before finish
`

const helpBugfix = `Manage bugfix branches across all modules.

Bugfix branches follow the same pattern as feature branches but signal
a bug fix rather than a new feature. Branched from develop.

Subcommands:
  start    Create bugfix/<name> from develop
  finish   Merge into develop and delete branch
  publish  Push bugfix to remote
  track    Track remote bugfix locally
  delete   Delete bugfix branch

Examples:
  gflow bugfix start login-validation
  gflow bugfix finish login-validation
`

const helpRelease = `Manage release branches across all modules.

Release branches prepare the codebase for a new production version.
gflow validates semver, optionally bumps manifests, and generates a
CHANGELOG entry on finish.

Subcommands:
  start    Create release/<version> from develop in all modules
  finish   Merge into main + develop, tag, push tags, bump CHANGELOG

Examples:
  gflow release start 2.0.0
  gflow release finish 2.0.0
  gflow release finish 2.0.0 --changelog --push-tags
  gflow release finish --bump minor --changelog   # auto-bump semver

Flags (finish):
  --bump patch|minor|major   Auto-increment version from latest tag
  --changelog                Generate CHANGELOG.md entry from commits
  --push-tags                Push tags to remote after finish (default: true)

Guards:
  - SemverGuard on start
  - RemoteSyncGuard on finish
`

const helpMerge = `Merge a branch into all modules.

Runs git merge <branch> [extra git flags] across every resolved module.
Useful for merging hotfixes into feature branches, or syncing any
branch that git-flow doesn't handle automatically.

Examples:
  gflow merge develop
  gflow merge main --no-ff
  gflow merge hotfix/1.2.1

  # Only specific modules:
  gflow merge develop -P api -P web

  # Interactive module picker:
  gflow merge develop -i

Flags:
  -i, --interactive    Pick modules via TUI
  --parallel           Run concurrently
  --dry-run            Print commands without executing
  --fail-fast          Stop on first error

Any extra arguments after the branch name are passed verbatim to git merge.
`

const helpStatus = `Show working tree status across all modules.

Runs git status --short in each module and aggregates results.
Modules with no changes are shown as clean.

Examples:
  gflow status
  gflow status --json           # machine-readable output
  gflow status -i               # interactive picker
  gflow status --parallel       # parallel (faster on many modules)
`

const helpPull = `Pull in all modules.

Runs git pull [remote] [branch] in each module.

Examples:
  gflow pull
  gflow pull origin
  gflow pull origin develop
  gflow pull --recurse-submodules    # also run submodule update after pull
  gflow pull --parallel

Flags:
  --recurse-submodules    Run git submodule update --init --recursive after pull
  --parallel              Run concurrently
  --fail-fast             Stop on first error
`

const helpPush = `Push in all modules.

Runs git push [remote] [branch] in each module.
Automatically sets upstream with --set-upstream if branch has no remote tracking.

Examples:
  gflow push
  gflow push origin
  gflow push origin feature/my-feature
  gflow push --parallel
`

const helpCheckout = `Checkout a branch in all modules.

Runs git checkout <branch> in each module.

Examples:
  gflow checkout develop
  gflow checkout main --recurse-submodules
  gflow checkout feature/my-feature -i    # pick modules interactively

Flags:
  --recurse-submodules    Run git submodule update --init --recursive after checkout
  -i, --interactive       Pick modules via TUI
`

const helpSubmodule = `Manage git submodule lifecycle.

Provides atomic operations for adding, removing, moving, and syncing
submodules — keeping .gitmodules consistent in a single command.

Subcommands:
  add <url> <path>     Add submodule and stage .gitmodules change
  remove <path>        deinit + git rm + clean .git/modules/<path>
  move <old> <new>     git mv + submodule sync
  sync                 Refresh URLs from .gitmodules (git submodule sync --recursive)
  drift                Show submodules whose HEAD ≠ parent-registered SHA

Examples:
  gflow submodule add https://github.com/org/lib libs/shared
  gflow submodule add https://github.com/org/lib libs/shared -b main
  gflow submodule remove libs/old
  gflow submodule move libs/shared packages/shared
  gflow submodule sync
  gflow submodule drift

Flags:
  -b, --branch    Branch to track (add subcommand only)
  --force         Skip confirmation on remove
  --dry-run       Print commands without executing

Tip: after add/remove, update MODULES array in gflow.yaml or .gflow.conf.
`

const helpWorktree = `Manage git worktrees across modules.

Worktrees let you work on multiple features in the same repo simultaneously
without stashing. Each worktree is an independent checkout of a branch.

Requires git ≥ 2.38 for stable submodule+worktree support.

Subcommands:
  add <branch> [--path dir]   Create worktree for branch across modules
  list                        Aggregate worktrees from all modules
  remove <path>               Remove worktree + prune stale entries
  switch <branch>             Print path of existing worktree (for cd)

Examples:
  gflow worktree add feature/my-feature
  gflow worktree add feature/my-feature --path /tmp/my-feature-wt
  gflow worktree list
  gflow worktree remove /tmp/my-feature-wt
  cd $(gflow worktree switch feature/my-feature)

Shell alias tip:
  gwt-switch() { cd "$(gflow worktree switch "$1")"; }
`

const helpProjects = `Manage global project registry.

Stores a list of project aliases → absolute paths in
~/.config/gflow/projects.yaml (respects $XDG_CONFIG_HOME).

Use gflow cd <alias> to jump between projects without typing full paths.

Subcommands:
  add <alias> [path]   Register a project (defaults to current directory)
  list                 Show all registered projects with branch + dirty status
  remove <alias>       Remove a project from registry
  recent [-n N]        Show N most recently used projects

Examples:
  gflow projects add backend
  gflow projects add infra ~/work/infra
  gflow projects list
  gflow projects recent -n 5
  gflow projects remove old-project

Shell setup (add to ~/.zshrc or ~/.bashrc):
  gflow-cd() { cd "$(gflow cd "$1")"; }
  alias gcd=gflow-cd

Then: gcd backend
`

const helpPR = `Manage pull requests across modules.

Wraps the gh CLI to create, inspect, and merge PRs for every module
that has the current branch published to a remote.

Requires: gh CLI (https://cli.github.com) authenticated with your provider.

Subcommands:
  create   Open a PR in each module with a published branch
  status   Show PR state, checks, and mergeability per module
  merge    Merge PRs across all modules using a consistent strategy

Examples:
  gflow pr create --title "feat: new feature" --base develop
  gflow pr create --draft
  gflow pr status
  gflow pr merge --strategy squash

Flags (create):
  --base string     Base branch (default: develop)
  --title string    PR title
  --body string     PR description body
  --draft           Create as draft PR

Flags (merge):
  --strategy string   merge | squash | rebase (default: merge)
`

const helpHistory = `Show the history of gflow operations.

Every mutating operation (feature finish, release finish, submodule add, etc.)
is recorded in .git/gflow/journal.jsonl with timestamps, module SHAs before
and after, and per-module outcomes.

Use gflow undo to revert the last recorded operation.

Examples:
  gflow history
  gflow history -n 5    # show last 5 entries

Flags:
  -n, --count int   Number of entries to show (default 20)
`

const helpUndo = `Revert the last recorded gflow operation.

Reads the last entry from .git/gflow/journal.jsonl and resets each
affected module to its pre-operation HEAD SHA via git reset --hard.

This covers the partial-failure case: if feature finish succeeded in
modules 1-4 but failed in module 5, undo resets all 5 back.

Examples:
  gflow undo
  gflow undo --module api    # undo only for the 'api' module
  gflow undo --force         # skip confirmation prompt
  gflow undo --dry-run       # preview what would be reset

Flags:
  --module string   Undo only for a specific module name
  --force           Skip confirmation
  --dry-run         Preview resets without executing
`

const helpDoctor = `Run diagnostic checks on the gflow environment.

Checks include:
  - git version (≥ 2.38 required for worktree+submodule support)
  - git-flow (AVH or community variant) installed and initialized
  - gflow config valid (.gflow.conf or gflow.yaml)
  - All declared modules exist and are git repositories
  - gh CLI available (required for gflow pr commands)
  - Shell completion installed
  - Submodule pointers in sync (no drift)

Examples:
  gflow doctor
  gflow doctor --json
`
