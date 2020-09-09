package output

import (
	"fmt"
	"io"

	"github.com/glebarez/padre/pkg/color"
)

// some often used strings
const (
	_LF = "\n"           // LF Line feed
	_CR = "\x1b\x5b2K\r" // Clear Line + CR Carret return
)

// Printer is the printing facility
type Printer struct {
	Stream         io.Writer // the ultimate stream to print into
	AvailableWidth int       // available terminal width
	cr             bool      // flag: caret return requested on next print (= print on same line please)
	prefix         *prefix   // current  prefix to use
}

// base internal print, everyone else must build upon this
func (p *Printer) print(s string) {
	fmt.Fprint(p.Stream, s)
}

func (p *Printer) Print(s string) {
	// CR debt ?
	if p.cr {
		p.print(_CR)
		p.cr = false
	}

	// prefix
	if p.prefix != nil {
		p.print(p.prefix.string())
	}

	// print the contents
	p.print(s)
}

// AddPrefix adds one more prefix to current printer
func (p *Printer) AddPrefix(s string) {
	p.prefix = newPrefix(s, p.prefix)
	p.AvailableWidth -= p.prefix.len
}

func (p *Printer) RemovePrefix() {
	p.AvailableWidth += p.prefix.len
	p.prefix = p.prefix.outterPrefix
}

func (p *Printer) Println(s string) {
	p.Print(s)
	p.print(_LF)

	// set flag that line was feeded
	if p.prefix != nil {
		p.prefix.setLF()
	}
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
}

func (p *Printer) Printcrf(format string, a ...interface{}) {
	p.Printcr(fmt.Sprintf(format, a...))
	p.cr = true
}

func (p *Printer) printWithPrefix(prefix, message string) {
	p.AddPrefix(prefix)
	p.Println(message)
	p.RemovePrefix()
}

func (p *Printer) Error(err error) {
	p.printWithPrefix(color.RedBold("[-]"), color.Red(err))
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
