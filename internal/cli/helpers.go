package cli

import (
	"errors"
	"fmt"

	"github.com/vmsfigueredo/gflow/internal/prompt"
)

// confirmPrompt asks a yes/no question in the terminal.
// Returns true if the user confirms. Returns false (no error) on abort.
func confirmPrompt(question string) (bool, error) {
	if !prompt.IsInteractive() {
		return false, fmt.Errorf("cannot confirm in non-interactive mode; use --force to bypass")
	}
	ok, err := prompt.ConfirmSummary(question)
	if errors.Is(err, prompt.ErrAborted) {
		return false, nil
	}
	return ok, err
}
