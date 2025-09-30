package codefmt

import (
	"fmt"
	"go/token"
)

// CodeError indicates where the error occurred in user's source code.
type CodeError struct {
	err  error
	pos  token.Pos
	end  token.Pos
	fset *token.FileSet
}

// Unwrap returns the underlying error.
func (e CodeError) Unwrap() error { return e.err }

// Pos returns the position where the error occurred. It may be invalid.
func (e CodeError) Pos() token.Pos { return e.pos }

// End returns the end position of the error. It may be invalid.
func (e CodeError) End() token.Pos { return e.end }

// Error implements the error interface. If pos is valid, the position is
// prepended to the error message.
func (e CodeError) Error() string {
	if e.err == nil {
		return ""
	}

	if !e.pos.IsValid() {
		return e.err.Error()
	}

	return fmt.Sprintf("%s: %s", FormatPosition(e.fset.Position(e.pos)), e.err.Error())
}

// Errorf formats an error message. The error will indicate the position in the
// source code if the position is valid.
func (f Formatter) Errorf(poser Poser, format string, args ...any) error {
	// Prevent wrapping error in args
	for _, arg := range args {
		if _, ok := arg.(error); ok {
			panic("CodeError cannot wrap error")
		}
	}

	var pos, end token.Pos
	if poser != nil {
		pos = poser.Pos()
		if ender, ok := poser.(Ender); ok {
			end = ender.End()
		}
	}

	args = f.wrapPrintfArgs(args)
	err := fmt.Errorf(format, args...)
	return &CodeError{err, pos, end, f.Fset}
}
