package errors

import (
	"errors"
	"fmt"
	"runtime"
)

// cockroachError is a lightweight approximation of github.com/cockroachdb/errors behaviour
// sufficient for this demo when the real module is unavailable.
type cockroachError struct {
	msg         string
	detail      string
	safeDetails []string
	stack       []uintptr
	cause       error
}

// Newf mimics errors.Newf by formatting the message and capturing stack frames.
func Newf(format string, args ...interface{}) error {
	return &cockroachError{
		msg:   fmt.Sprintf(format, args...),
		stack: captureStack(),
	}
}

// Wrap adds context around an error while retaining the cause and stack.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &cockroachError{
		msg:   msg,
		cause: err,
		stack: captureStack(),
	}
}

// WithDetail attaches structured detail to the error tree.
func WithDetail(err error, detail string) error {
	if err == nil {
		return nil
	}
	ce := ensureCockroach(err)
	ce.detail = detail
	return ce
}

// WithSafeDetails annotates the error with safe-to-log strings and stack.
func WithSafeDetails(err error, msg string, details ...string) error {
	if err == nil {
		return nil
	}
	return &cockroachError{
		msg:         msg,
		cause:       err,
		stack:       captureStack(),
		safeDetails: append([]string(nil), details...),
	}
}

// GetSafeDetails walks the chain searching for safe details.
func GetSafeDetails(err error) []string {
	if err == nil {
		return nil
	}
	type safeDetailer interface {
		SafeDetails() []string
	}
	if sd, ok := err.(safeDetailer); ok {
		if details := sd.SafeDetails(); len(details) > 0 {
			return append([]string(nil), details...)
		}
	}
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		return GetSafeDetails(unwrapped)
	}
	return nil
}

func (e *cockroachError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.cause)
	}
	return e.msg
}

func (e *cockroachError) Unwrap() error {
	return e.cause
}

func (e *cockroachError) SafeDetails() []string {
	return append([]string(nil), e.safeDetails...)
}

func (e *cockroachError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s\n", e.msg)
			if e.detail != "" {
				fmt.Fprintf(s, "detail: %s\n", e.detail)
			}
			for _, pc := range e.stack {
				fn := runtime.FuncForPC(pc)
				if fn == nil {
					continue
				}
				file, line := fn.FileLine(pc)
				fmt.Fprintf(s, "  %s:%d\n", file, line)
			}
			if e.cause != nil {
				fmt.Fprintf(s, "caused by: %+v\n", e.cause)
			}
			return
		}
		fallthrough
	default:
		fmt.Fprintf(s, "%s", e.Error())
	}
}

func ensureCockroach(err error) *cockroachError {
	if ce, ok := err.(*cockroachError); ok {
		copy := *ce
		return &copy
	}
	return &cockroachError{
		msg:   err.Error(),
		cause: err,
		stack: captureStack(),
	}
}

func captureStack() []uintptr {
	const depth = 16
	pcs := make([]uintptr, depth)
	n := runtime.Callers(3, pcs)
	return pcs[:n]
}
