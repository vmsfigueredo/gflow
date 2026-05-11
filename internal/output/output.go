package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/vmsfigueredo/gflow/internal/executor"
	"github.com/vmsfigueredo/gflow/internal/module"
)

// HeaderFlags carries display hints for PrintHeader.
type HeaderFlags struct {
	Parallel bool
	DryRun   bool
}

// symbols holds the icon set in use (unicode or ASCII fallback).
type symbols struct {
	ok      string
	fail    string
	dryRun  string
	skip    string
	header  string
	bullet  string
}

var (
	plain bool // true → no color, ASCII symbols

	symUnicode = symbols{ok: "✓", fail: "✗", dryRun: "~", skip: "·", header: "▸", bullet: "•"}
	symASCII   = symbols{ok: "+", fail: "x", dryRun: "~", skip: ".", header: ">", bullet: "*"}

	sym symbols

	colorOK     func(a ...interface{}) string
	colorFail   func(a ...interface{}) string
	colorWarn   func(a ...interface{}) string
	colorBold   func(a ...interface{}) string
	colorDim    func(a ...interface{}) string
	colorHeader func(a ...interface{}) string
)

// Init configures output style. Must be called before any Print* function.
// noColor: --no-color flag. Reads NO_COLOR env and isatty internally.
func Init(noColor bool) {
	isTTY := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	noColorEnv := os.Getenv("NO_COLOR") != ""
	plain = noColor || noColorEnv || !isTTY

	if plain {
		sym = symASCII
		id := func(a ...interface{}) string { return fmt.Sprint(a...) }
		colorOK, colorFail, colorWarn, colorBold, colorDim, colorHeader = id, id, id, id, id, id
		color.NoColor = true
		return
	}

	sym = symUnicode
	colorOK = color.New(color.FgGreen).SprintFunc()
	colorFail = color.New(color.FgRed).SprintFunc()
	colorWarn = color.New(color.FgYellow).SprintFunc()
	colorBold = color.New(color.Bold).SprintFunc()
	colorDim = color.New(color.FgHiBlack).SprintFunc()
	colorHeader = color.New(color.Bold, color.FgCyan).SprintFunc()
}

func init() {
	// Sensible defaults before Init is called (e.g., in tests).
	Init(false)
}

// PrintHeader prints the command banner before results.
func PrintHeader(cmd, target string, modCount int, flags HeaderFlags) {
	if plain && modCount == 1 {
		return
	}
	title := cmd
	if target != "" {
		title += " " + target
	}
	chips := fmt.Sprintf("%d module", modCount)
	if modCount != 1 {
		chips += "s"
	}
	if flags.Parallel {
		chips += "  " + sym.bullet + "  parallel"
	}
	if flags.DryRun {
		chips += "  " + sym.bullet + "  dry-run"
	}
	fmt.Printf("\n%s %s  %s  %s\n\n", colorHeader(sym.header), colorBold(title), colorDim(sym.bullet), colorDim(chips))
}

// Print writes results to stdout (human or JSON).
func Print(results []executor.Result, asJSON bool) error {
	if asJSON {
		return printJSON(results)
	}

	// Column width = longest module name.
	maxLen := 0
	for _, r := range results {
		if n := len(r.Module.Display); n > maxLen {
			maxLen = n
		}
	}

	failed, skipped, ok := 0, 0, 0
	var totalDur time.Duration
	for _, r := range results {
		printResult(r, maxLen)
		totalDur += r.Duration
		switch r.Status {
		case executor.StatusOK:
			ok++
		case executor.StatusError:
			failed++
		case executor.StatusSkip:
			skipped++
		}
	}

	printSummary(ok, failed, skipped, totalDur)

	if failed > 0 {
		return fmt.Errorf("%d operation(s) failed", failed)
	}
	return nil
}

func printResult(r executor.Result, nameWidth int) {
	name := fmt.Sprintf("%-*s", nameWidth, r.Module.Display)
	dur := fmtDur(r.Duration)

	switch r.Status {
	case executor.StatusOK:
		fmt.Printf("  %s %s  [%s]   %s\n",
			colorOK(sym.ok), colorBold(name), r.Action, colorDim(dur))
		if r.Output != "" {
			printIndented(r.Output, 6)
		}
	case executor.StatusError:
		fmt.Printf("  %s %s  [%s]   %s\n",
			colorFail(sym.fail), colorBold(name), r.Action, colorDim(dur))
		if r.Err != nil {
			fmt.Printf("      %s\n", colorFail(r.Err.Error()))
		}
		if r.Output != "" {
			printIndented(r.Output, 6)
		}
	case executor.StatusDryRun:
		fmt.Printf("  %s %s  [%s]   %s  %s\n",
			colorWarn(sym.dryRun), colorBold(name), r.Action, colorDim(dur), colorDim("(dry-run)"))
		if r.Output != "" {
			printIndented(r.Output, 6)
		}
	case executor.StatusSkip:
		fmt.Printf("  %s %s  [%s]\n",
			colorDim(sym.skip), colorDim(name), r.Action)
	}
}

func printIndented(s string, indent int) {
	pad := strings.Repeat(" ", indent)
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for _, l := range lines {
		fmt.Printf("%s%s\n", pad, l)
	}
}

func printSummary(ok, failed, skipped int, total time.Duration) {
	parts := []string{}
	if ok > 0 {
		parts = append(parts, colorOK(fmt.Sprintf("%d ok", ok)))
	}
	if failed > 0 {
		parts = append(parts, colorFail(fmt.Sprintf("%d failed", failed)))
	}
	if skipped > 0 {
		parts = append(parts, colorDim(fmt.Sprintf("%d skipped", skipped)))
	}
	if len(parts) == 0 {
		return
	}
	sep := colorDim("  " + sym.bullet + "  ")
	line := strings.Join(parts, sep)
	durStr := colorDim("total " + fmtDur(total))
	fmt.Printf("\n  %s   %s\n\n", line, durStr)
}

func fmtDur(d time.Duration) string {
	if d == 0 {
		return ""
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%02ds", m, s)
}

type jsonResult struct {
	Module   string `json:"module"`
	Display  string `json:"display"`
	Action   string `json:"action"`
	Status   string `json:"status"`
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
}

func printJSON(results []executor.Result) error {
	rows := make([]jsonResult, len(results))
	for i, r := range results {
		jr := jsonResult{
			Module:   r.Module.Name,
			Display:  r.Module.Display,
			Action:   r.Action,
			Status:   string(r.Status),
			Output:   r.Output,
			Duration: r.Duration.String(),
		}
		if r.Err != nil {
			jr.Error = r.Err.Error()
		}
		rows[i] = jr
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}

// PrintList prints the module list (for `gflow list`).
func PrintList(mods []*module.Module, asJSON bool) error {
	if asJSON {
		type row struct {
			Name string `json:"name"`
			Path string `json:"path"`
			Root bool   `json:"root"`
		}
		rows := make([]row, len(mods))
		for i, m := range mods {
			rows[i] = row{Name: m.Display, Path: m.Path, Root: m.Root}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}
	for _, m := range mods {
		tag := ""
		if m.Root {
			tag = colorDim("  (root)")
		}
		fmt.Printf("  %s%s\n", colorBold(m.Display), tag)
	}
	return nil
}

// PrintAliases prints the alias table (for `gflow aliases`).
func PrintAliases(aliases map[string][]string, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(aliases)
	}
	for k, paths := range aliases {
		fmt.Printf("  %s  %s  %s\n", colorBold(k), colorDim(sym.bullet), strings.Join(paths, ", "))
	}
	return nil
}

const doctorLabelWidth = 32

// PrintDoctorCheck writes a single doctor check line with aligned columns.
func PrintDoctorCheck(label string, ok bool, detail string) {
	var icon string
	if ok {
		icon = colorOK(sym.ok)
	} else {
		icon = colorFail(sym.fail)
	}
	padded := fmt.Sprintf("%-*s", doctorLabelWidth, label)
	line := fmt.Sprintf("  %s %s", icon, padded)
	if detail != "" {
		line += colorDim("  —  ") + detail
	}
	fmt.Println(line)
}

// ── Help formatting ──────────────────────────────────────────────────────────

// PrintHelpHeader prints the top description line.
func PrintHelpHeader(short string) {
	fmt.Printf("  %s\n", colorBold(short))
}

// HelpUsage returns a styled usage string.
func HelpUsage(usage string) string {
	return colorDim(usage)
}

// HelpInlineCode returns a styled inline code snippet for help text.
func HelpInlineCode(s string) string {
	return colorBold(s)
}

// PrintHelpSection prints one grouped section of commands.
func PrintHelpSection(title string, names []string, byName map[string]*cobra.Command) {
	fmt.Printf("  %s\n", colorHeader("  "+title))
	maxLen := 0
	for _, n := range names {
		if len(n) > maxLen {
			maxLen = len(n)
		}
	}
	for _, n := range names {
		c, ok := byName[n]
		if !ok {
			continue
		}
		fmt.Printf("    %s  %s\n",
			colorBold(fmt.Sprintf("%-*s", maxLen, n)),
			colorDim(c.Short),
		)
	}
	fmt.Println()
}

// HelpFlagsSection returns a formatted flags block for help output.
func HelpFlagsSection(cmd *cobra.Command) string {
	var sb strings.Builder
	sb.WriteString("  " + colorHeader("  Flags") + "\n")

	printFlag := func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		name := "--" + f.Name
		if f.ShorthandDeprecated == "" && f.Shorthand != "" {
			name = "-" + f.Shorthand + ", " + name
		}
		val := ""
		if f.Value.Type() != "bool" {
			val = " " + f.Value.Type()
		}
		sb.WriteString(fmt.Sprintf("    %s  %s\n",
			colorBold(fmt.Sprintf("%-28s", name+val)),
			colorDim(f.Usage),
		))
	}

	// -h/--help always first
	sb.WriteString(fmt.Sprintf("    %s  %s\n",
		colorBold(fmt.Sprintf("%-28s", "-h, --help")),
		colorDim("help for "+cmd.Name()),
	))
	cmd.PersistentFlags().VisitAll(printFlag)
	sb.WriteString("\n")
	return sb.String()
}

// Infof prints an informational message to stdout.
func Infof(format string, args ...interface{}) {
	fmt.Printf("  "+format+"\n", args...)
}

// Successf prints a success message to stdout.
func Successf(format string, args ...interface{}) {
	fmt.Printf("  %s "+format+"\n", append([]interface{}{colorOK(sym.ok)}, args...)...)
}

// Warnf prints a warning message to stdout.
func Warnf(format string, args ...interface{}) {
	fmt.Printf("  %s "+format+"\n", append([]interface{}{colorWarn("!")}, args...)...)
}

// Errorf prints an error message to stderr.
func Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "  %s "+format+"\n", append([]interface{}{colorFail(sym.fail)}, args...)...)
}

// DoctorSummary prints the footer after all doctor checks.
func DoctorSummary(passed, failed int) {
	parts := []string{}
	if passed > 0 {
		parts = append(parts, colorOK(fmt.Sprintf("%d passed", passed)))
	}
	if failed > 0 {
		parts = append(parts, colorFail(fmt.Sprintf("%d failed", failed)))
	}
	if len(parts) == 0 {
		return
	}
	sep := colorDim("  " + sym.bullet + "  ")
	fmt.Printf("\n  %s\n\n", strings.Join(parts, sep))
}
