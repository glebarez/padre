package output

import (
	"strings"

	"github.com/glebarez/padre/pkg/color"
)

const (
	empty = ``
	space = ` `
)

// represents a current prefix
// the prefix allows for contexted printing
// the prefixes can be nested using outterPrefix attribute
// the top-most prefix has outterPrefix equal to nil
type prefix struct {
	prefix       string  // the prefix to be output
	indent       string  // indent to iutput on second+ lines of multiline outputs
	len          int     // length of prefix and indent
	lineFeeded   bool    // flag: line feeded (=true when first line was already output)
	outterPrefix *prefix // pointer to outter parent prefix
}

// renders prefix as string
func (p *prefix) string() string {
	var s string

	// form own prefix as string
	if p.lineFeeded {
		s = p.indent
	} else {
		s = p.prefix + space
	}

	// add outter prefix (if any)
	if p.outterPrefix == nil {
		return s
	}
	return p.outterPrefix.string() + s
}

// sets lineFeeded flag
func (p *prefix) setLF() {
	p.lineFeeded = true
	if p.outterPrefix != nil {
		p.outterPrefix.setLF()
	}
}

// creates child prefix for current prefix
// returns the created child
func (p *prefix) add(s string) *prefix {
	child := &prefix{
		prefix:       s,
		indent:       strings.Repeat(space, color.TrueLen(s)+1),
		len:          len(s),
		outterPrefix: p,
	}

	return child
}
