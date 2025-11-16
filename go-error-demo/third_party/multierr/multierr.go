package multierr

import (
	"errors"
	"strings"
)

type multiError []error

// Append combines the provided errors. Nil values are ignored.
func Append(left error, right error) error {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	errs := flatten(left)
	errs = append(errs, flatten(right)...)
	return multiError(errs)
}

// Combine joins an arbitrary list of errors.
func Combine(errs ...error) error {
	var out error
	for _, err := range errs {
		out = Append(out, err)
	}
	return out
}

// Errors returns the underlying slice of errors, similar to the upstream package.
func Errors(err error) []error {
	if err == nil {
		return nil
	}
	if me, ok := err.(multiError); ok {
		return append([]error(nil), me...)
	}
	if uw := errors.Unwrap(err); uw != nil {
		return []error{err}
	}
	return []error{err}
}

func (m multiError) Error() string {
	var b strings.Builder
	for i, err := range m {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

func (m multiError) Unwrap() []error {
	return append([]error(nil), m...)
}

func flatten(err error) []error {
	if err == nil {
		return nil
	}
	if me, ok := err.(multiError); ok {
		return append([]error(nil), me...)
	}
	return []error{err}
}
