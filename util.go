package main

import (
	"fmt"

	"github.com/fatih/color"
)

var stderr = color.Error

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

func (p *argErrors) error(flag string, err error) {
	e := fmt.Errorf("Parameter %s: %w", flag, err)
	p.errors = append(p.errors, e)
}

func (p *argErrors) errorf(flag string, format string, a ...interface{}) {
	e := fmt.Errorf("Parameter %s: %s", fmt.Sprintf(format, a...))
	p.errors = append(p.errors, e)
}

func (p *argErrors) warningf(flag string, format string, a ...interface{}) {
	w := fmt.Sprintf("Parameter %s: %s", fmt.Sprintf(format, a...))
	p.warnings = append(p.warnings, w)
}
