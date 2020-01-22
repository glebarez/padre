package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var logo = `
 ▄▄ •        ▄▄▄· ▄▄▄· ·▄▄▄▄  ·▄▄▄▄   ▄· ▄▌
▐█ ▀ ▪▪     ▐█ ▄█▐█ ▀█ ██▪ ██ ██▪ ██ ▐█▪██▌
▄█ ▀█▄ ▄█▀▄  ██▀·▄█▀▀█ ▐█· ▐█▌▐█· ▐█▌▐█▌▐█▪
▐█▄▪▐█▐█▌.▐▌▐█▪·•▐█ ▪▐▌██. ██ ██. ██  ▐█▀·.
·▀▀▀▀  ▀█▄▀▪.▀    ▀  ▀ ▀▀▀▀▀• ▀▀▀▀▀•   ▀ • 
`

func printLogo() {
	indent := "                 "
	cyan := color.New(color.FgCyan)

	for _, s := range strings.Split(logo, "\n") {
		fmt.Fprintf(color.Error, cyan.Sprintf("%s%s\n", indent, s))
	}
}
