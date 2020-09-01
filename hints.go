package main

import (
	"github.com/glebarez/padre/pkg/config"
)

// flag wrapper
func _f(f string) string {
	return `(` + color.GreenBold(`-`+f) + ` option)`
}

var hint = struct {
	omitBlockLen     string
	omitErrPattern   string
	setErrPattern    string
	lowerConnections string
	checkEncoding    string
	checkInput       string
}{
	omitBlockLen:     `omit ` + _f(`b`) + `  for automatic detection of block length`,
	omitErrPattern:   `omit ` + _f(`err`) + ` for automatic fingerprinting of HTTP responses`,
	setErrPattern:    `specify error pattern manually with ` + _f(`err`),
	lowerConnections: `server might be overwhelmed or rate-limiting you requests. try lowering concurrency using ` + _f(`p`),
	checkEncoding:    `check that encoding ` + _f(`e`) + ` and replacement rules ` + _f(`r`) + ` are set properly`,
	checkInput:       `check that INPUT is properly formatted`,
}

func makeDetectionHints(*config.Config) []string {
	// hint intro
	intro := `if you believe target is vulnerable, try following:`
	li := color.CyanBold(`> `)
	hints := []string{intro}

	// block length
	if *config.blockLen != 0 {
		hints = append(hints, li+hint.omitBlockLen)
	} else {
		// error pattern
		if *config.paddingErrorPattern != "" {
			hints = append(hints, li+hint.omitErrPattern)
		} else {
			hints = append(hints, li+hint.setErrPattern)
		}
	}

	// concurrency
	if *config.parallel > 10 {
		hints = append(hints, li+hint.lowerConnections)
	}
	return hints
}
