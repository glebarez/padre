package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/glebarez/padre/pkg/color"
)

const (
	// LF Line feed
	LF = "\n"

	// CR Carret return
	CR = "\x1b\x5b2K\r"
)

type Printer struct {
	Stream io.Writer
}

func (p *Printer) Print(a ...interface{}) {
	fmt.Fprint(p.Stream, a...)
}

func (p *Printer) Println(a ...interface{}) {
	fmt.Fprintln(p.Stream, a...)
}

func (p *Printer) Printcr(a ...interface{}) {
	p.Print(fmt.Sprint(a...), CR)
}

func (p *Printer) printWithPrefix(prefix, message string) {
	message = strings.TrimSpace(message)
	lines := strings.Split(message, LF)
	for i, line := range lines {
		if i != 0 {
			prefix = strings.Repeat(" ", len(color.StripColor(prefix)))
		}
		p.Println(fmt.Sprintf("%s %s", prefix, line))
	}
}

func (p *Printer) Error(err error) {
	p.printWithPrefix(color.RedBold("[-]"), color.Red(err))

	// // print hints if available
	// var ewh *errors.ErrWithHints
	// if errors.As(err, &ewh) {
	// 	p.printHint(strings.Join(ewh.hints, LF))
	// }
}

func (p *Printer) Errorf(format string, a ...interface{}) {
	p.Error(fmt.Errorf(format, a...))
}

func (p *Printer) Hint(format string, a ...interface{}) {
	p.printWithPrefix(color.CyanBold("[hint]"), fmt.Sprintf(format, a...))
}

func (p *Printer) Warning(format string, a ...interface{}) {
	p.printWithPrefix(color.YellowBold("[!]"), fmt.Sprintf(format, a...))
}

func (p *Printer) Success(format string, a ...interface{}) {
	p.printWithPrefix(color.GreenBold("[+]"), fmt.Sprintf(format, a...))
}

func (p *Printer) Info(format string, a ...interface{}) {
	p.printWithPrefix(color.CyanBold("[i]"), fmt.Sprintf(format, a...))
}

func (p *Printer) Action(s string) {
	p.Printcr(color.Yellow(s))
}
