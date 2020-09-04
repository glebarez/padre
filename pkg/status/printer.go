package out

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/fatih/color"
)

var (
	colorMatcher *regexp.Regexp
	leftover     string
	outputStream = color.Error
)

/* status-aware printer */
func Print(s string) {
	if currentStatus != nil {
		currentStatus.print(s, false)
	} else {
		_, err := fmt.Fprintln(outputStream, leftover+s)
		if err != nil {
			log.Fatal(err)
		}
	}
}

/* fatal printer */
func die(e error) {
	printError(e)
	os.Exit(1)
}
