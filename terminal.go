package main

import (
	"github.com/nsf/termbox-go"
)

func consoleWidth() int {
	if err := termbox.Init(); err != nil {
		return -1
	}
	w, _ := termbox.Size()
	termbox.Close()
	return w
}
