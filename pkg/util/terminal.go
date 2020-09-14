package util

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/nsf/termbox-go"
)

// TerminalWidth determines width of current terminal in characters
func TerminalWidth() (int, error) {
	if err := termbox.Init(); err != nil {
		return 0, err
	}
	w, _ := termbox.Size()
	termbox.Close()
	// decrease length by 1 for safety
	// windows CMD sometimes needs this
	return w - 1, nil
}

// IsTerminal checks whether file is a terminal
func IsTerminal(file *os.File) bool {
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
