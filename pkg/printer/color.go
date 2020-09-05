package printer

import (
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var colorMatcher *regexp.Regexp

func init() {
	// override the standard decision on No-color mode
	color.NoColor = os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()))
	// matcher for coloring terminal sequences
	colorMatcher = regexp.MustCompile("\033\\[.*?m")
}

/* coloring stringers */
var (
	red         = color.New(color.FgRed).SprintFunc()
	bold        = color.New(color.Bold).SprintFunc()
	Yellow      = color.New(color.FgYellow).SprintFunc()
	redBold     = color.New(color.FgRed, color.Bold).SprintFunc()
	CyanBold    = color.New(color.FgCyan, color.Bold).SprintFunc()
	cyan        = color.New(color.FgCyan).SprintFunc()
	GreenBold   = color.New(color.FgGreen, color.Bold).SprintFunc()
	green       = color.New(color.FgGreen).SprintFunc()
	hiGreenBold = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	underline   = color.New(color.Underline).SprintFunc()
	yellowBold  = color.New(color.FgYellow, color.Bold).SprintFunc()
)

// strips terminal color controls  from a string
func StripColor(s string) string {
	return colorMatcher.ReplaceAllString(s, "")
}
