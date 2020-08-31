package errors

// ErrWithHints - error with hints attached
type ErrWithHints struct {
	err   error
	hints []string
}

func (e *ErrWithHints) Error() string {
	return e.err.Error()
}

// NewErrWithHints - factory
func NewErrWithHints(err error, hints ...string) *ErrWithHints {
	return &ErrWithHints{
		err:   err,
		hints: hints,
	}
}

// FatalError ...
type FatalError string

func (e FatalError) Error() string { return string(e) }
