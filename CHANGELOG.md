# Changelog

## [Unreleased]

### Added
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
