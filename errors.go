package goerr

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// New creates a new error with message
func New(msg string) *Error {
	return newError(msg)
}

// Wrap creates a new Error and add message
func Wrap(cause error, msg ...interface{}) *Error {
	if len(msg) > 0 {
		var newMsgs []string
		for _, m := range msg {
			newMsgs = append(newMsgs, fmt.Sprintf("%v", m))
		}
		err := newError(strings.Join(newMsgs, " "))
		err.cause = cause
		return err
	}

	err, ok := cause.(*Error)
	if !ok {
		err = newError("wrapped error")
		err.cause = cause
		return err
	}

	return err
}

// Error is error interface for deepalert to handle related variables
type Error struct {
	msg    string
	code   string
	st     *stack
	cause  error
	values map[string]interface{}
}

func newError(msg string) *Error {
	return &Error{
		msg:    msg,
		st:     callers(),
		values: make(map[string]interface{}),
	}
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
// - %v, %s, %q: formated message
// - %+v: formated message with stack trace
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
func (x *Error) With(key string, value interface{}) *Error {
	x.values[key] = value
	return x
}

// Is returns true if target is goerr.Error and Error.code of two errors are matched. It's for errors.Is. If Error.code is empty, it always returns false.
func (x *Error) Is(target error) bool {
	var err *Error
	if errors.As(target, &err) {
		if x.code != "" && x.code == err.code {
			return true
		}
	}

	return false
}

// Code sets string to check equality in Error.IS()
func (x *Error) Code(code string) *Error {
	x.code = code
	return x
}

// Error.Wrap creates a new Error and copy message and code to new one.
func (x *Error) Wrap(cause error) *Error {
	err := newError(x.msg)
	err.cause = cause
	err.code = x.code
	return err
}

// Values returns map of key and value that is set by With. All wrapped goerr.Error key and values will be merged. Key and values of wrapped error is overwritten by upper goerr.Error.
func (x *Error) Values() map[string]interface{} {
	var values map[string]interface{}

	if cause := x.Unwrap(); cause != nil {
		if err, ok := cause.(*Error); ok {
			values = err.Values()
		}
	}

	if values == nil {
		values = make(map[string]interface{})
	}

	for key, value := range x.values {
		values[key] = value
	}

	return values
}
