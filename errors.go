package goerr

import (
	"errors"
	"fmt"
	"io"

	"log/slog"

	"github.com/google/uuid"
)

type Option func(*Error)

// Value sets key and value to the error
func Value(key string, value any) Option {
	return func(err *Error) {
		err.values[key] = value
	}
}

// V is alias of Value
func V(key string, value any) Option {
	return Value(key, value)
}

// Tag sets tag to the error
func Tag(t tag) Option {
	return func(err *Error) {
		err.tags[t] = struct{}{}
	}
}

// T is alias of Tag
func T(t tag) Option {
	return Tag(t)
}

// New creates a new error with message
func New(msg string, options ...Option) *Error {
	err := newError(options...)
	err.msg = msg
	return err
}

// Wrap creates a new Error and add message.
func Wrap(cause error, msg string, options ...Option) *Error {
	err := newError(options...)
	err.msg = msg
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

// Values returns map of key and value that is set by With. All wrapped goerr.Error key and values will be merged. Key and values of wrapped error is overwritten by upper goerr.Error.
func Values(err error) map[string]any {
	if e := Unwrap(err); e != nil {
		return e.Values()
	}

	return nil
}

// Tags returns list of tags that is set by WithTags. All wrapped goerr.Error tags will be merged. Tags of wrapped error is overwritten by upper goerr.Error.
func Tags(err error) []string {
	if e := Unwrap(err); e != nil {
		return e.Tags()
	}

	return nil
}

// HasTag returns true if the error has the tag.
func HasTag(err error, tag tag) bool {
	if e := Unwrap(err); e != nil {
		return e.HasTag(tag)
	}

	return false
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

func newError(options ...Option) *Error {
	e := &Error{
		st:     callers(),
		values: make(values),
		id:     uuid.New().String(),
		tags:   make(tags),
	}

	for _, opt := range options {
		opt(e)
	}

	return e
}

func (x *Error) copy(dst *Error, options ...Option) {
	dst.msg = x.msg
	dst.id = x.id
	dst.cause = x.cause

	dst.tags = x.tags.clone()
	dst.values = x.values.clone()

	for _, opt := range options {
		opt(dst)
	}
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
			_, _ = io.WriteString(s, "\n")

			if len(x.values) > 0 {
				_, _ = io.WriteString(s, "\nValues:\n")
				for k, v := range x.values {
					_, _ = io.WriteString(s, fmt.Sprintf("  %s: %v\n", k, v))
				}
				_, _ = io.WriteString(s, "\n")
			}
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
func (x *Error) Wrap(cause error, options ...Option) *Error {
	err := newError()
	x.copy(err, options...)
	err.cause = cause
	return err
}

// Values returns map of key and value that is set by With. All wrapped goerr.Error key and values will be merged. Key and values of wrapped error is overwritten by upper goerr.Error.
func (x *Error) Values() map[string]any {
	values := x.mergedValues()

	for key, value := range x.values {
		values[key] = value
	}

	return values
}

func (x *Error) mergedValues() values {
	merged := make(values)

	if cause := x.Unwrap(); cause != nil {
		if err := Unwrap(cause); err != nil {
			merged = err.mergedValues()
		}
	}

	for key, value := range x.values {
		merged[key] = value
	}

	return merged
}

// Tags returns list of tags that is set by WithTags. All wrapped goerr.Error tags will be merged. Tags of wrapped error is overwritten by upper goerr.Error.
func (x *Error) Tags() []string {
	tags := x.mergedTags()

	for tag := range x.tags {
		tags[tag] = struct{}{}
	}

	tagList := make([]string, 0, len(tags))
	for tag := range tags {
		tagList = append(tagList, tag.value)
	}

	return tagList
}

func (x *Error) mergedTags() tags {
	merged := make(tags)

	if cause := x.Unwrap(); cause != nil {
		if err := Unwrap(cause); err != nil {
			merged = err.mergedTags()
		}
	}

	for tag := range x.tags {
		merged[tag] = struct{}{}
	}

	return merged
}

// LogValue returns slog.Value for structured logging. It's implementation of slog.LogValuer.
// https://pkg.go.dev/log/slog#LogValuer
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

	var tags []string
	for tag := range x.tags {
		tags = append(tags, tag.value)
	}
	attrs = append(attrs, slog.Any("tags", tags))

	var stacktrace any
	var traces []string
	for _, st := range x.StackTrace() {
		traces = append(traces, fmt.Sprintf("%s:%d %s", st.getFilePath(), st.getLineNumber(), st.getFunctionName()))
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
