package errors

import "github.com/glebarez/padre/pkg/color"

// flag wrapper
func _f(f string) string {
	return `(` + color.GreenBold(`-`+f) + ` option)`
}

var Hint = struct {
	omitBlockLen     string
	omitErrPattern   string
	setErrPattern    string
	LowerConnections string
	checkEncoding    string
	checkInput       string
}{
	omitBlockLen:     `omit ` + _f(`b`) + `  for automatic detection of block length`,
	omitErrPattern:   `omit ` + _f(`err`) + ` for automatic fingerprinting of HTTP responses`,
	setErrPattern:    `specify error pattern manually with ` + _f(`err`),
	LowerConnections: `server might be overwhelmed or rate-limiting you requests. try lowering concurrency using ` + _f(`p`),
	checkEncoding:    `check that encoding ` + _f(`e`) + ` and replacement rules ` + _f(`r`) + ` are set properly`,
	checkInput:       `check that INPUT is properly formatted`,
}
