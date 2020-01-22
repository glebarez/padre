package main

import "github.com/fatih/color"

var red = color.New(color.FgRed).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var redBold = color.New(color.FgRed, color.Bold).SprintFunc()
var cyanBold = color.New(color.FgCyan, color.Bold).SprintFunc()
var greenBold = color.New(color.FgGreen, color.Bold).SprintFunc()
