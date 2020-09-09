package main

import (
	"github.com/glebarez/padre/pkg/color"
	"github.com/glebarez/padre/pkg/output"
)

// flag wrapper
func _f(f string) string {
	return `(` + color.GreenBold(`-`+f) + ` option)`
}

// hint texts
var (
	omitBlockLen     = `omit ` + _f(`b`) + `  for automatic detection of block length`
	omitErrPattern   = `omit ` + _f(`err`) + ` for automatic fingerprinting of HTTP responses`
	setErrPattern    = `specify error pattern manually with ` + _f(`err`)
	lowerConnections = `server might be overwhelmed or rate-limiting you requests. try lowering concurrency using ` + _f(`p`)
	checkEncoding    = `check that encoding ` + _f(`e`) + ` and replacement rules ` + _f(`r`) + ` are set properly`
	checkInput       = `check that INPUT is properly formatted`
)

// make hints for obvious reasons
func makeDetectionHints(args *Args, p *output.Printer) {
	// hints intro
	p.AddPrefix(color.CyanBold("[hints]"), true)
	defer p.RemovePrefix()

	p.Println(`if you believe target is vulnerable, try following:`)

	// list hints
	p.AddPrefix(color.CyanBold(`> `), false)
	defer p.RemovePrefix()

	// block length
	if *args.BlockLen != 0 {
		p.Println(omitBlockLen)
	} else {
		// error pattern
		if *args.PaddingErrorPattern != "" {
			p.Println(omitErrPattern)
		} else {
			p.Println(setErrPattern)
		}
	}

	// concurrency
	if *args.Parallel > 10 {
		p.Println(lowerConnections)
	}
}
