package output

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

	empty = ""
	space = " "
)

type Printer struct {
	Stream        io.Writer
	TerminalWidth int
	cr            bool
	prefix        string
	indent        string
	lineFeeded    bool
}

func (p *Printer) print(s string) {
	fmt.Fprint(p.Stream, s)
}

func (p *Printer) Print(s string) {
	// CR debt ?
	if p.cr {
		p.print(CR)
		p.cr = false
	}

	// prefix logic (if set)
	if p.prefix != empty {
		if p.lineFeeded {
			p.print(p.indent)
		} else {
			p.print(p.prefix)
			p.print(space)
		}
	}

	// base print
	p.print(s)
}

func (p *Printer) SetPrefix(prefix string) {
	p.prefix = prefix
	p.indent = strings.Repeat(space, color.TrueLen(prefix)+1)
	p.lineFeeded = false
	p.TerminalWidth -= len(p.indent)
}

func (p *Printer) ResetPrefix() {
	p.TerminalWidth += len(p.indent)
	p.SetPrefix(empty)
}

func (p *Printer) Println(s string) {
	p.Print(s)
	p.print(LF)

	// set flag that line was feeded
	p.lineFeeded = true
}

func (p *Printer) Printcr(s string) {
	p.Print(s)
	p.cr = true
}

func (p *Printer) Printf(format string, a ...interface{}) {
	p.Print(fmt.Sprintf(format, a...))
}

func (p *Printer) Printlnf(format string, a ...interface{}) {
	p.Println(fmt.Sprintf(format, a...))
	p.Print(LF)
}

func (p *Printer) Printcrf(format string, a ...interface{}) {
	p.Printcr(fmt.Sprintf(format, a...))
	p.cr = true
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
