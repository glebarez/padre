package main

import "fmt"

type argErrors struct {
	errors   []error
	warnings []string
}

func newArgErrors() *argErrors {
	return &argErrors{
		errors:   make([]error, 0),
		warnings: make([]string, 0),
	}
}

func (p *argErrors) flagError(flag string, err error) {
	e := fmt.Errorf("Parameter %s: %w", flag, err)
	p.errors = append(p.errors, e)
}

func (p *argErrors) flagErrorf(flag string, format string, a ...interface{}) {
	e := fmt.Errorf("Parameter %s: %s", flag, fmt.Sprintf(format, a...))
	p.errors = append(p.errors, e)
}

func (p *argErrors) flagWarningf(flag string, format string, a ...interface{}) {
	w := fmt.Sprintf("Parameter %s: %s", flag, fmt.Sprintf(format, a...))
	p.warnings = append(p.warnings, w)
}

func (p *argErrors) warningf(format string, a ...interface{}) {
	w := fmt.Sprintf(format, a...)
	p.warnings = append(p.warnings, w)
}
