# Changelog

## [Unreleased]

### Added
- feat(main): `gflow update` command with auto-detect install source (homebrew vs binary), GitHub Releases check, SHA256 verification, atomic replace
- feat(main): `--verbose / -v` flag streaming raw git output to stderr
- feat(main): goreleaser config with changelog, checksum sha256, windows/arm64 ignore, homebrew_casks
- chore(main): ignore root `gflow` binary artifact in .gitignore
- chore(main): remove homebrew_casks from goreleaser until tap repo exists

### Changed
- chore(main): `.goreleaser.yaml` migrated deprecated `format`/`brews` fields to `formats`/`homebrew_casks`
- fix(main): `pull`/`push` arg parsing — single arg now sets branch, not remote
- feat(main): `git status` output enriched with branch name and clean/modified state
- feat(develop): complete Go rewrite of gflow with polished terminal UX
- feat(develop): interactive module picker for flow, pull, push, checkout, status, commit commands
- feat(develop): `feature update` subcommand integrating develop via merge or rebase
- feat(develop): `--names api=foo,web=bar` flag for per-module branch names in flow ops
- feat(develop): `--recurse-submodules` flag for pull and checkout
- feat(develop): `--project/-P` flag now repeatable (union of multiple aliases)
- feat(develop): submodule, worktree, projects, cd, history, undo, pr, tag, merge commands
- feat(develop): operation journal tracking refs before/after each gitflow op
- feat(develop): `RunOpFn` indirection in gitflow/ops for test overriding
- feat(develop): output helpers Infof, Successf, Warnf, Errorf
- feat(develop): GH API client (internal/gh) for PR creation
- feat(develop): release changelog and manifest packages
- feat(develop): registry package for module registry
- feat(develop): auto push --set-upstream on "no upstream branch" error

### Fixed
- fix(develop): .gflow.conf parser now supports multi-line array syntax MODULES=(\n"a"\n"b"\n)
- fix(main): remove `--json` from `gh pr create` (unsupported flag), fetch PR data via `gh pr view` after creation
- fix(main): remove release body (changelog) from `gflow update` output to prevent pipe corruption on `curl | sh` installs

### Changed
- feat(main): `gflow tag list` now groups tags by major.minor, highlights latest tag
