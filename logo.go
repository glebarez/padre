package main

import (
	"fmt"
	"os"
	"strings"
)

var logo = `
 ▄▄ •        ▄▄▄· ▄▄▄· ·▄▄▄▄  ·▄▄▄▄   ▄· ▄▌
▐█ ▀ ▪▪     ▐█ ▄█▐█ ▀█ ██▪ ██ ██▪ ██ ▐█▪██▌
▄█ ▀█▄ ▄█▀▄  ██▀·▄█▀▀█ ▐█· ▐█▌▐█· ▐█▌▐█▌▐█▪
▐█▄▪▐█▐█▌.▐▌▐█▪·•▐█ ▪▐▌██. ██ ██. ██  ▐█▀·.
·▀▀▀▀  ▀█▄▀▪.▀    ▀  ▀ ▀▀▀▀▀• ▀▀▀▀▀•   ▀ • 
`

func Logo() {
	indent := "                 "

	for _, s := range strings.Split(logo, "\n") {
		fmt.Fprintf(os.Stderr, "%s%s\n", indent, s)
	}
}
