package printer

import (
	"fmt"
	"io"
	"strings"
)

const (
	// LF Line feed
	LF = "\n"

	// CR Carret return
	CR = "\x1b\x5b2K\r"
)

type Printer struct {
	stream io.Writer
}

func (p *Printer) Print(a ...interface{}) {
	fmt.Fprint(p.stream, a)
}

func (p *Printer) Println(a ...interface{}) {
	fmt.Fprintln(p.stream, a)
}

func (p *Printer) Printcr(a ...interface{}) {
	p.Print(fmt.Sprint(a...), CR)
}

func (p *Printer) printWithPrefix(prefix, message string) {
	message = strings.TrimSpace(message)
	lines := strings.Split(message, LF)
	for i, line := range lines {
		if i != 0 {
			prefix = strings.Repeat(" ", len(stripColor(prefix)))
		}
		p.Println(fmt.Sprintf("%s %s", prefix, line))
	}
}

func (p *Printer) Error(err error) {
	p.printWithPrefix(redBold("[-]"), red(err))

	// // print hints if available
	// var ewh *errors.ErrWithHints
	// if errors.As(err, &ewh) {
	// 	p.printHint(strings.Join(ewh.hints, LF))
	// }
}

func (p *Printer) Hint(message string) {
	p.printWithPrefix(CyanBold("[hint]"), message)
}

func (p *Printer) Warning(message string) {
	p.printWithPrefix(yellowBold("[!]"), message)
}

func (p *Printer) Success(message string) {
	p.printWithPrefix(GreenBold("[+]"), message)
}

func (p *Printer) Info(message string) {
	p.printWithPrefix(CyanBold("[i]"), message)
}

func (p *Printer) Action(s string) {
	p.Printcr(Yellow(s))
}
