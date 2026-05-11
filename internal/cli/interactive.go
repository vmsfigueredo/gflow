package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/vmsfigueredo/gflow/internal/git"
	"github.com/vmsfigueredo/gflow/internal/module"
	"github.com/vmsfigueredo/gflow/internal/prompt"
)

// errSilentAbort signals a clean user-initiated abort (no further error output needed).
var errSilentAbort = errors.New("")

// pickInteractive shows the module picker and returns the selected subset.
// Returns (nil, nil) when not in a TTY or --json is set — caller uses full mods list.
// Returns (nil, errSilentAbort) on user cancel.
func pickInteractive(ctx context.Context, mods []*module.Module) ([]*module.Module, error) {
	if !prompt.IsInteractive() || flagJSON {
		return nil, nil
	}
	hints := git.BatchCurrentBranches(ctx, mods)
	picked, err := prompt.PickModules(mods, hints)
	if err != nil {
		if errors.Is(err, prompt.ErrAborted) {
			fmt.Println("Aborted.")
			return nil, errSilentAbort
		}
		return nil, err
	}
	if len(picked) == 0 {
		fmt.Println("No modules selected. Aborted.")
		return nil, errSilentAbort
	}
	return picked, nil
}
