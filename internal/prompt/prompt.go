package prompt

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// ErrNotTTY is returned when prompts are invoked in a non-interactive context.
var ErrNotTTY = errors.New("prompt: stdin is not a tty")

// ErrAborted is returned when the user cancels an interactive prompt.
var ErrAborted = errors.New("prompt: user aborted")

// IsInteractive reports whether it is safe to show interactive prompts.
// Respects GFLOW_NO_INTERACTIVE env for scripted use in otherwise-TTY contexts.
func IsInteractive() bool {
	if os.Getenv("GFLOW_NO_INTERACTIVE") != "" {
		return false
	}
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// AskName prompts for a branch name. def pre-populates the text field.
func AskName(branchType, op, def string) (string, error) {
	if !IsInteractive() {
		return "", ErrNotTTY
	}
	var name string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("%s %s — branch name:", branchType, op)).
				Value(&name).
				Placeholder(def).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("name cannot be empty")
					}
					return nil
				}),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrAborted
		}
		return "", err
	}
	if name == "" {
		name = def
	}
	return strings.TrimSpace(name), nil
}

// PickModules shows a multi-select of all modules, all pre-checked.
// hints maps module.Name to its current branch string for display.
func PickModules(mods []*module.Module, hints map[string]string) ([]*module.Module, error) {
	if !IsInteractive() {
		return nil, ErrNotTTY
	}

	// Build huh options — all pre-selected.
	opts := make([]huh.Option[string], len(mods))
	selected := make([]string, len(mods))
	for i, m := range mods {
		label := m.Display
		if b, ok := hints[m.Name]; ok && b != "" {
			label = fmt.Sprintf("%-20s  (%s)", m.Display, b)
		}
		opts[i] = huh.NewOption(label, m.Name)
		selected[i] = m.Name
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select modules:").
				Options(opts...).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrAborted
		}
		return nil, err
	}

	// Map back to *module.Module in original order.
	selSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selSet[s] = true
	}
	var picked []*module.Module
	for _, m := range mods {
		if selSet[m.Name] {
			picked = append(picked, m)
		}
	}
	return picked, nil
}

// AskPerModuleName prompts for a branch name for each module individually.
// defaults maps module.Name to a pre-filled default string.
// An empty answer for a module omits it from the returned map (= skip).
func AskPerModuleName(branchType string, mods []*module.Module, defaults map[string]string) (map[string]string, error) {
	if !IsInteractive() {
		return nil, ErrNotTTY
	}

	result := make(map[string]string, len(mods))
	for _, m := range mods {
		def := ""
		if defaults != nil {
			def = defaults[m.Name]
		}

		var name string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("%s name for [%s]:", branchType, m.Display)).
					Description("Leave empty to skip this module.").
					Placeholder(def).
					Value(&name),
			),
		)
		if err := form.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil, ErrAborted
			}
			return nil, err
		}
		if name == "" {
			name = def
		}
		name = strings.TrimSpace(name)
		if name != "" {
			result[m.Name] = name
		}
	}
	return result, nil
}

// AskTagMessage prompts for an annotated tag message for release finish.
// def is the default message shown as placeholder.
func AskTagMessage(version, def string) (string, error) {
	if !IsInteractive() {
		return "", ErrNotTTY
	}
	var msg string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Tag message for release %s:", version)).
				Description("Leave empty to use default.").
				Placeholder(def).
				Value(&msg),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrAborted
		}
		return "", err
	}
	if strings.TrimSpace(msg) == "" {
		return def, nil
	}
	return strings.TrimSpace(msg), nil
}

// PickBranch shows a single-select of branch names and returns the chosen one.
// def is the branch to pre-select (cursor); ignored if not in the list.
func PickBranch(title string, branches []string, def string) (string, error) {
	if !IsInteractive() {
		return "", ErrNotTTY
	}

	opts := make([]huh.Option[string], len(branches))
	for i, b := range branches {
		opts[i] = huh.NewOption(b, b)
	}

	selected := def
	if selected == "" && len(branches) > 0 {
		selected = branches[0]
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(opts...).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrAborted
		}
		return "", err
	}
	return selected, nil
}

// ConfirmSummary shows the given summary text and asks yes/no.
// Returns ErrAborted when user declines.
func ConfirmSummary(summary string) (bool, error) {
	if !IsInteractive() {
		return false, ErrNotTTY
	}
	var ok bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(summary).
				Value(&ok).
				Affirmative("Yes").
				Negative("No"),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}
		return false, err
	}
	return ok, nil
}
