package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

func init() {
	// override the standard decision on No-color mode
	color.NoColor = os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()))
}

var red = color.New(color.FgRed).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var redBold = color.New(color.FgRed, color.Bold).SprintFunc()
var cyanBold = color.New(color.FgCyan, color.Bold).SprintFunc()
var greenBold = color.New(color.FgGreen, color.Bold).SprintFunc()

func printError(e error) {
	fmt.Fprintf(color.Error, redBold(e.Error()+"\n"))
}
