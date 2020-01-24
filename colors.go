package main

import (
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
var yellow = color.New(color.FgHiYellow).SprintFunc()
var redBold = color.New(color.FgRed, color.Bold).SprintFunc()
var cyanBold = color.New(color.FgCyan, color.Bold).SprintFunc()
var greenBold = color.New(color.FgGreen, color.Bold).SprintFunc()
var underline = color.New(color.Underline).SprintFunc()
var yellowBold = color.New(color.FgYellow, color.Bold).SprintFunc()
