package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
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
func print(s string) {
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
func printError(err error) {
	// print error message
	printWithPrefix(redBold("[-]"), red(err))

	// print hints if available
	var ewh *errWithHints
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
func printWarning(message string) {
	printWithPrefix(yellowBold("[!]"), message)
}

/* success */
func printSuccess(message string) {
	printWithPrefix(greenBold("[+]"), message)
}

/*  hint */
func printHint(message string) {
	printWithPrefix(cyanBold("[hint]"), message)
}

/* info */
func printInfo(message string) {
	printWithPrefix(cyanBold("[i]"), message)
}

/* fatal printer */
func die(e error) {
	printError(e)
	os.Exit(1)
}

/* coloring stringers */
var red = color.New(color.FgRed).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var redBold = color.New(color.FgRed, color.Bold).SprintFunc()
var cyanBold = color.New(color.FgCyan, color.Bold).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()
var greenBold = color.New(color.FgGreen, color.Bold).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var hiGreenBold = color.New(color.FgHiGreen, color.Bold).SprintFunc()
var underline = color.New(color.Underline).SprintFunc()
var yellowBold = color.New(color.FgYellow, color.Bold).SprintFunc()

// error with hints
type errWithHints struct {
	err   error
	hints []string
}

func (e *errWithHints) Error() string {
	return e.err.Error()
}

func newErrWithHints(err error, hints ...string) *errWithHints {
	return &errWithHints{
		err:   err,
		hints: hints,
	}
}
