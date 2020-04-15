package main

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/nsf/termbox-go"
)

/* determine width of current terminal */
func terminalWidth() int {
	if err := termbox.Init(); err != nil {
		return -1
	}
	w, _ := termbox.Size()
	termbox.Close()
	return w
}

/* is terminal? */
func isTerminal(file *os.File) bool {
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
