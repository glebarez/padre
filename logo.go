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
	// * no logo in narrow terminal
	if config.termWidth < 46 {
		return
	}
	indent := strings.Repeat(" ", (config.termWidth-44)/2)
	cyan := color.New(color.FgCyan)

	for _, s := range strings.Split(logo, "\n") {
		fmt.Fprintf(color.Error, cyan.Sprintf("%s%s\n", indent, s))
	}
}
