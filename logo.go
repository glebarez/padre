package main

import "fmt"

var logo = `
 ▄▄ •        ▄▄▄· ▄▄▄· ·▄▄▄▄  ·▄▄▄▄   ▄· ▄▌
▐█ ▀ ▪▪     ▐█ ▄█▐█ ▀█ ██▪ ██ ██▪ ██ ▐█▪██▌
▄█ ▀█▄ ▄█▀▄  ██▀·▄█▀▀█ ▐█· ▐█▌▐█· ▐█▌▐█▌▐█▪
▐█▄▪▐█▐█▌.▐▌▐█▪·•▐█ ▪▐▌██. ██ ██. ██  ▐█▀·.
·▀▀▀▀  ▀█▄▀▪.▀    ▀  ▀ ▀▀▀▀▀• ▀▀▀▀▀•   ▀ • 
`

func Logo() {
	fmt.Println(logo)
}