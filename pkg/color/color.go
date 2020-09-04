package color

import "github.com/fatih/color"

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
