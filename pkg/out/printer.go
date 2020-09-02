package out

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/glebarez/padre/pkg/errors"
	"github.com/mattn/go-isatty"
)

var (
	colorMatcher *regexp.Regexp
	leftover     string
	outputStream = color.Error
)

func init() {
	// override the standard decision on No-color mode
	color.NoColor = os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()))

	// matcher for coloring terminal sequences
	colorMatcher = regexp.MustCompile("\033\\[.*?m")
}

// strip terminal color controls  from a string
func stripColor(s string) string {
	return colorMatcher.ReplaceAllString(s, "")
}

/* status-aware printer */
func Print(s string) {
	if currentStatus != nil {
		currentStatus.print(s, false)
	} else {
		_, err := fmt.Fprintln(outputStream, leftover+s)
		if err != nil {
			log.Fatal(err)
		}
	}
}

/* print with prefix, supports multiline messages */
func printWithPrefix(prefix string, message string) {
	lines := strings.Split(message, "\n")
	for i, line := range lines {
		if i != 0 {
			prefix = strings.Repeat(" ", len(stripColor(prefix)))
		}
		print(fmt.Sprintf("%s %s", prefix, line))
	}
}

/* error printer */
func PrintError(err error) {
	// print error message
	printWithPrefix(redBold("[-]"), red(err))

	// print hints if available
	var ewh *errors.ErrWithHints
	if errors.As(err, &ewh) {
		printHint(strings.Join(ewh.hints, "\n"))
	}
}

/* action printer */
func printAction(s string) {
	if currentStatus != nil {
		currentStatus.print(yellow(s), true)
	} else {
		_, err := fmt.Fprint(outputStream, leftover+yellow(s))
		if err != nil {
			log.Fatal(err)
		}
		leftover = "\x1b\x5b2K\r" // clear line + caret return
	}
}

/* warning  */
func PrintWarning(message string) {
	printWithPrefix(yellowBold("[!]"), message)
}

/* success */
func PrintSuccess(message string) {
	printWithPrefix(greenBold("[+]"), message)
}

/*  hint */
func printHint(message string) {
	printWithPrefix(cyanBold("[hint]"), message)
}

/* info */
func PrintInfo(message string) {
	printWithPrefix(cyanBold("[i]"), message)
}

/* fatal printer */
func die(e error) {
	printError(e)
	os.Exit(1)
}
