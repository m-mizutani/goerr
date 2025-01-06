package goerr

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"log/slog"

	"github.com/google/uuid"
)

// New creates a new error with message
func New(format string, args ...any) *Error {
	err := newError()
	err.msg = fmt.Sprintf(format, args...)
	return err
}

func toWrapMessage(msgs []any) string {
	var newMsgs []string
	for _, m := range msgs {
		newMsgs = append(newMsgs, fmt.Sprintf("%v", m))
	}
	return strings.Join(newMsgs, " ")
}

// Wrap creates a new Error and add message.
func Wrap(cause error, msg ...any) *Error {
	err := newError()
	err.msg = toWrapMessage(msg)
	err.cause = cause

	return err
}

// Wrapf creates a new Error and add message. The error message is formatted by fmt.Sprintf.
func Wrapf(cause error, format string, args ...any) *Error {
	err := newError()
	err.msg = fmt.Sprintf(format, args...)
	err.cause = cause
	return err
}

// Unwrap returns unwrapped goerr.Error from err by errors.As. If no goerr.Error, returns nil
// NOTE: Do not receive error interface. It causes typed-nil problem.
//
//	var err error = goerr.New("error")
//	if err != nil { // always true
func Unwrap(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return nil
}

type values map[string]any

func (x values) clone() values {
	newValues := make(values)
	for key, value := range x {
		newValues[key] = value
	}
	return newValues
}

// Error is error interface for deepalert to handle related variables
type Error struct {
	msg    string
	id     string
	st     *stack
	cause  error
	values values
	tags   tags
}

func newError() *Error {
	return &Error{
		st:     callers(),
		values: make(values),
		id:     uuid.New().String(),
		tags:   make(tags),
	}
}

func (x *Error) copy(dst *Error) {
	dst.msg = x.msg
	dst.id = x.id
	dst.cause = x.cause
	dst.tags = x.tags.clone()
	dst.values = x.values.clone()
	// st (stacktrace) is not copied
}

// Printable returns printable object
func (x *Error) Printable() *Printable {
	e := &Printable{
		Message:    x.msg,
		ID:         x.id,
		StackTrace: x.Stacks(),
		Values:     make(map[string]any),
	}
	for k, v := range x.values {
		e.Values[k] = v
	}
	for tag := range x.tags {
		e.Tags = append(e.Tags, tag.value)
	}

	if cause := Unwrap(x.cause); cause != nil {
		e.Cause = cause.Printable()
	} else if x.cause != nil {
		e.Cause = x.cause.Error()
	}
	return e
}

type Printable struct {
	Message    string         `json:"message"`
	ID         string         `json:"id"`
	StackTrace []*Stack       `json:"stacktrace"`
	Cause      any            `json:"cause"`
	Values     map[string]any `json:"values"`
	Tags       []string       `json:"tags"`
}

// Error returns error message for error interface
func (x *Error) Error() string {
	if x.cause == nil {
		return x.msg
	}

	return fmt.Sprintf("%s: %v", x.msg, x.cause.Error())
}

// Format returns:
// - %v, %s, %q: formatted message
// - %+v: formatted message with stack trace
func (x *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, x.Error())
			var c *Error
			for c = x; c.Unwrap() != nil; {
				cause, ok := c.Unwrap().(*Error)
				if !ok {
					break
				}
				c = cause
			}
			c.st.Format(s, verb)
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

// Unstack trims stack trace by 1. It can be used for internal helper or utility functions.
func (x *Error) Unstack() *Error {
	x.st = unstack(x.st, 1)
	return x
}

// UnstackN trims stack trace by n. It can be used for internal helper or utility functions.
func (x *Error) UnstackN(n int) *Error {
	x.st = unstack(x.st, n)
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

func (x *Error) LogValue() slog.Value {
	if x == nil {
		return slog.AnyValue(nil)
	}

	attrs := []slog.Attr{
		slog.String("message", x.msg),
	}
	var values []any
	for k, v := range x.values {
		values = append(values, slog.Any(k, v))
	}
	attrs = append(attrs, slog.Group("values", values...))

	var stacktrace any
	var traces []string
	for _, st := range x.StackTrace() {
		traces = append(traces, fmt.Sprintf("%s:%d %s", st.file(), st.line(), st.name()))
	}
	stacktrace = traces

	attrs = append(attrs, slog.Any("stacktrace", stacktrace))

	if x.cause != nil {
		var errAttr slog.Attr
		if lv, ok := x.cause.(slog.LogValuer); ok {
			errAttr = slog.Any("cause", lv.LogValue())
		} else {
			errAttr = slog.Any("cause", x.cause)
		}
		attrs = append(attrs, errAttr)
	}

	return slog.GroupValue(attrs...)
}
