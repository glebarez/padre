package printer

import (
	"fmt"
	"io"
	"strings"
)

const (
	// LF Linefeed
	LF = "\n"

	// CR Carret return
	CR = "\x1b\x5b2K\r"
)

type Printer struct {
	stream   io.Writer
}

func (p *Printer) Print(a ...interface{}) {
	fmt.Fprint(p.stream, a)
}

func (p *Printer) Println(a ...interface{}) {
	fmt.Fprintln(p.stream, a)
}

func (p *Printer) Printcr(a ...interface{}) {
	p.Print(a..., CR)
}

func (p *Print) printWithPrefix(prefix, message string) {
	message = strings.TrimSpace(message)
	lines := strings.Split(message, LF)
	for i, line := range lines {
		if i != 0 {
			prefix = strings.Repeat(" ", len(stripColor(prefix)))
		}
		p.Println(fmt.Sprintf("%s %s", prefix, line))
	}
}

func (p *Print) PrintError(err error) {
	// print error message
	p.printWithPrefix(redBold("[-]"), red(err))

	// print hints if available
	var ewh *errors.ErrWithHints
	if errors.As(err, &ewh) {
		p.printHint(strings.Join(ewh.hints, LF))
	}
}

/*  hint */
func (p *Print) printHint(message string) {
	p.printWithPrefix(cyanBold("[hint]"), message)
}

/* warning  */
func (p *Print) PrintWarning(message string) {
	p.printWithPrefix(yellowBold("[!]"), message)
}

/* success */
func (p *Print) PrintSuccess(message string) {
	p.printWithPrefix(greenBold("[+]"), message)
}

/* info */
func (p *Print) PrintInfo(message string) {
	p.printWithPrefix(cyanBold("[i]"), message)
}

/* action printer */
func (p *Print) PrintAction(s string) {
	if currentStatus != nil {
		currentStatus.print(yellow(s), true)
	} else {
		_, err := fmt.Fprint(outputStream, leftover+yellow(s))
		if err != nil {
			log.Fatal(err)
		}
		leftover = "\x1b\x5b2K\r" // clear line + caret return
	}
}