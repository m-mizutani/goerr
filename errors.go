package goerr

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

// New creates a new error with message
func New(format string, args ...any) *Error {
	err := newError()
	err.msg = fmt.Sprintf(format, args...)
	return err
}

// Wrap creates a new Error and add message
func Wrap(cause error, msg ...any) *Error {
	err := newError()

	if len(msg) > 0 {
		var newMsgs []string
		for _, m := range msg {
			newMsgs = append(newMsgs, fmt.Sprintf("%v", m))
		}
		err.msg = strings.Join(newMsgs, " ")
	}

	err.cause = cause

	return err
}

// Unwrap returns unwrapped goerr.Error from err by errors.As. If no goerr.Error, returns nil
func Unwrap(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return nil
}

// Error is error interface for deepalert to handle related variables
type Error struct {
	msg    string
	id     string
	st     *stack
	cause  error
	values map[string]any
}

func newError() *Error {
	return &Error{
		st:     callers(),
		values: make(map[string]any),
		id:     uuid.New().String(),
	}
}

func (x *Error) copy(dst *Error) {
	dst.msg = x.msg
	dst.id = x.id
	dst.cause = x.cause
	for k, v := range x.values {
		dst.values[k] = v
	}
	// st (stacktrace) is not copied
}

// Printable returns printable object
func (x *Error) Printable() *printable {
	e := &printable{
		Message:    x.msg,
		ID:         x.id,
		StackTrace: x.Stacks(),
		Cause:      x.cause,
		Values:     make(map[string]any),
	}
	for k, v := range x.values {
		e.Values[k] = v
	}
	return e
}

type printable struct {
	Message    string         `json:"message"`
	ID         string         `json:"id"`
	StackTrace []*Stack       `json:"stacktrace"`
	Cause      error          `json:"cause"`
	Values     map[string]any `json:"values"`
}

// Error returns error message for error interface
func (x *Error) Error() string {
	s := x.msg
	cause := x.cause
	for i := 0; i < 16; i++ {
		if cause == nil {
			break
		}

		s = fmt.Sprintf("%s: %v", s, cause.Error())
		type errorUnwrap interface {
			Unwrap() error
		}

		unwrapable, ok := cause.(errorUnwrap)
		if !ok {
			break
		}

		cause = unwrapable.Unwrap()
	}

	return s
}

// Format returns:
// - %v, %s, %q: formatted message
// - %+v: formatted message with stack trace
func (x *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, x.Error())
			x.st.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, x.Error())
	case 'q':
		fmt.Fprintf(s, "%q", x.Error())
	}
}

// Unwrap returns *fundamental of github.com/pkg/errors
func (x *Error) Unwrap() error {
	return x.cause
}

// With adds key and value related to the error event
func (x *Error) With(key string, value any) *Error {
	x.values[key] = value
	return x
}

// Is returns true if target is goerr.Error and Error.id of two errors are matched. It's for errors.Is. If Error.id is empty, it always returns false.
func (x *Error) Is(target error) bool {
	var err *Error
	if errors.As(target, &err) {
		if x.id != "" && x.id == err.id {
			return true
		}
	}

	return x == target
}

// ID sets string to check equality in Error.IS()
func (x *Error) ID(id string) *Error {
	x.id = id
	return x
}

// Wrap creates a new Error and copy message and id to new one.
func (x *Error) Wrap(cause error) *Error {
	err := newError()
	x.copy(err)
	err.cause = cause
	return err
}

// Values returns map of key and value that is set by With. All wrapped goerr.Error key and values will be merged. Key and values of wrapped error is overwritten by upper goerr.Error.
func (x *Error) Values() map[string]any {
	var values map[string]any

	if cause := x.Unwrap(); cause != nil {
		if err, ok := cause.(*Error); ok {
			values = err.Values()
		}
	}

	if values == nil {
		values = make(map[string]any)
	}

	for key, value := range x.values {
		values[key] = value
	}

	return values
}
