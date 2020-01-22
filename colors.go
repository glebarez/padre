package main

import "github.com/fatih/color"

var red = color.New(color.FgRed).SprintFunc()
var redBold = color.New(color.FgRed, color.Bold).SprintFunc()
var cyanBold = color.New(color.FgCyan, color.Bold).SprintFunc()
var greenBold = color.New(color.FgGreen, color.Bold).SprintFunc()
