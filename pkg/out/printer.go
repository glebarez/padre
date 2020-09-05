package out

import (
	"os"

	"github.com/fatih/color"
)

var (
	leftover     string
	outputStream = color.Error
)

/* fatal printer */
func die(e error) {
	printError(e)
	os.Exit(1)
}
