package goerr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"log/slog"
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

// ID sets an error ID for Is() comparison.
// Empty string ("") is treated as an invalid ID and will not be used for comparison.
// When ID is set, errors.Is() will compare errors by ID instead of pointer equality.
func ID(id string) Option {
	return func(err *Error) {
		err.id = id
	}
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
	// Check for Errors type first using AsErrors
	if errs := AsErrors(err); errs != nil {
		return errs.HasTag(tag)
	}

	// Check for Error type using Unwrap
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
	msg         string
	id          string // Default is empty string (""). Empty string is treated as invalid ID and will not be used for Is() comparison
	st          *stack
	cause       error
	values      values         // String-based values
	typedValues map[string]any // Type-safe values
	tags        tags
}

func newError(options ...Option) *Error {
	e := &Error{
		st:          callers(),
		values:      make(values),
		typedValues: make(map[string]any),
		id:          "", // Default to empty string. Empty string is treated as invalid ID
		tags:        make(tags),
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

	// Clone typed values
	dst.typedValues = make(map[string]any)
	for key, value := range x.typedValues {
		dst.typedValues[key] = value
	}

	for _, opt := range options {
		opt(dst)
	}
	// st (stacktrace) is not copied
}

// Printable returns printable object
func (x *Error) Printable() *Printable {
	e := &Printable{
		Message:     x.msg,
		ID:          x.id,
		StackTrace:  x.Stacks(),
		Values:      x.Values(),      // Use Values() to get merged string-based values from wrapped errors
		TypedValues: x.TypedValues(), // Use TypedValues() to get merged typed values from wrapped errors
		Tags:        x.Tags(),        // Use Tags() to get merged tags from wrapped errors
	}

	if cause := Unwrap(x.cause); cause != nil {
		e.Cause = cause.Printable()
	} else if x.cause != nil {
		e.Cause = x.cause.Error()
	}
	return e
}

type Printable struct {
	Message     string         `json:"message"`
	ID          string         `json:"id"`
	StackTrace  []*Stack       `json:"stacktrace"`
	Cause       any            `json:"cause"`
	Values      map[string]any `json:"values"`
	TypedValues map[string]any `json:"typed_values"`
	Tags        []string       `json:"tags"`
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

			// Use merged values from entire error chain
			mergedValues := x.Values()
			if len(mergedValues) > 0 {
				_, _ = io.WriteString(s, "\nValues:\n")
				// Sort keys for predictable output
				keys := make([]string, 0, len(mergedValues))
				for k := range mergedValues {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					_, _ = io.WriteString(s, fmt.Sprintf("  %s: %v\n", k, mergedValues[k]))
				}
				_, _ = io.WriteString(s, "\n")
			}

			// Use merged typed values from entire error chain
			mergedTypedValues := x.TypedValues()
			if len(mergedTypedValues) > 0 {
				_, _ = io.WriteString(s, "\nTyped Values:\n")
				// Sort keys for predictable output
				keys := make([]string, 0, len(mergedTypedValues))
				for k := range mergedTypedValues {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					_, _ = io.WriteString(s, fmt.Sprintf("  %s: %v\n", k, mergedTypedValues[k]))
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

// Is returns true if the target error matches this error. It's for errors.Is.
// If both errors have IDs set via goerr.ID() option (non-empty), they are compared by ID.
// Otherwise, pointer equality is used for comparison.
// Empty ID values (default) are not used for comparison.
func (x *Error) Is(target error) bool {
	var err *Error
	if errors.As(target, &err) {
		if x.id != "" && x.id == err.id {
			return true
		}
	}

	return x == target
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
	return x.mergedValues()
}

// TypedValues returns map of key and value that is set by TypedValue. All wrapped goerr.Error typed key and values will be merged. Key and values of wrapped error is overwritten by upper goerr.Error.
func (x *Error) TypedValues() map[string]any {
	return x.mergedTypedValues()
}

func (x *Error) mergedValues() values {
	merged := make(values)

	if cause := x.Unwrap(); cause != nil {
		if err := Unwrap(cause); err != nil {
			merged = err.mergedValues()
		}
	}

	// Merge string-based values
	for key, value := range x.values {
		merged[key] = value
	}

	return merged
}

func (x *Error) mergedTypedValues() map[string]any {
	merged := make(map[string]any)

	if cause := x.Unwrap(); cause != nil {
		if err := Unwrap(cause); err != nil {
			merged = err.mergedTypedValues()
		}
	}

	// Merge typed values
	for key, value := range x.typedValues {
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

	var typedValues []any
	for k, v := range x.typedValues {
		typedValues = append(typedValues, slog.Any(k, v))
	}
	attrs = append(attrs, slog.Group("typed_values", typedValues...))

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

// MarshalJSON implements json.Marshaler interface for Error type.
// It provides comprehensive JSON serialization including message, ID,
// stack trace, values, tags, and cause information.
func (x *Error) MarshalJSON() ([]byte, error) {
	if x == nil {
		return []byte("null"), nil
	}
	return json.Marshal(x.Printable())
}
