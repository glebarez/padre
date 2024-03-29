package color

import (
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var colorMatcher *regexp.Regexp

var Error = color.Error

func init() {
	// override the standard decision on No-color mode
	color.NoColor = os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()))
	// matcher for coloring terminal sequences
	colorMatcher = regexp.MustCompile("\033\\[.*?m")
}

/* coloring stringers */
var (
	Red         = color.New(color.FgRed).SprintFunc()
	Bold        = color.New(color.Bold).SprintFunc()
	Yellow      = color.New(color.FgYellow).SprintFunc()
	RedBold     = color.New(color.FgRed, color.Bold).SprintFunc()
	CyanBold    = color.New(color.FgCyan, color.Bold).SprintFunc()
	Cyan        = color.New(color.FgCyan).SprintFunc()
	GreenBold   = color.New(color.FgGreen, color.Bold).SprintFunc()
	Green       = color.New(color.FgGreen).SprintFunc()
	HiGreenBold = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	Underline   = color.New(color.Underline).SprintFunc()
	YellowBold  = color.New(color.FgYellow, color.Bold).SprintFunc()
)

// StripColor - strips ANSI color control characters from a string
func StripColor(s string) string {
	return colorMatcher.ReplaceAllString(s, "")
}

// TrueLen returns true length of a colorized string in characters
func TrueLen(s string) int {
	return len(StripColor(s))
}
