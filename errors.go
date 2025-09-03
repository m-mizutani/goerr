package goerr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
)

// Errors represents multiple errors as a single error
type Errors struct {
	errs []error
}

// Error implements error interface
func (x *Errors) Error() string {
	if x == nil || len(x.errs) == 0 {
		return ""
	}

	if len(x.errs) == 1 {
		return x.errs[0].Error()
	}

	var messages []string
	for _, err := range x.errs {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "\n")
}

// Unwrap returns all wrapped errors for Go 1.20+ multiple errors support
func (x *Errors) Unwrap() []error {
	if x == nil || len(x.errs) == 0 {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]error, len(x.errs))
	copy(result, x.errs)
	return result
}

// Is checks if any wrapped error matches target
func (x *Errors) Is(target error) bool {
	if x == nil {
		return false
	}

	for _, err := range x.errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// As finds the first error that matches target type
func (x *Errors) As(target any) bool {
	if x == nil {
		return false
	}

	for _, err := range x.errs {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

// Join creates a new Errors by combining multiple errors
func Join(errs ...error) *Errors {
	filtered := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			filtered = append(filtered, err)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	return &Errors{
		errs: filtered,
	}
}

// Append adds errors to existing Errors (inspired by go-multierror)
// If base is nil, creates a new Errors. Flattens nested Errors.
func Append(base *Errors, errs ...error) *Errors {
	if len(errs) == 0 {
		return base
	}

	// Create new Errors if base is nil (go-multierror pattern)
	if base == nil {
		base = &Errors{
			errs: make([]error, 0),
		}
	}

	for _, err := range errs {
		if err == nil {
			continue
		}

		// Flatten nested Errors (go-multierror pattern)
		if nestedErrs, ok := err.(*Errors); ok {
			base.errs = append(base.errs, nestedErrs.errs...)
		} else {
			base.errs = append(base.errs, err)
		}
	}

	return base
}

// AsErrors extracts goerr.Errors from err by errors.As. If no goerr.Errors, returns nil
// Complementary to goerr.Unwrap() which only extracts goerr.Error
func AsErrors(err error) *Errors {
	var e *Errors
	if errors.As(err, &e) {
		return e
	}
	return nil
}

// IsEmpty returns true if no errors are contained
func (x *Errors) IsEmpty() bool {
	return x == nil || len(x.errs) == 0
}

// Len returns the number of wrapped errors
func (x *Errors) Len() int {
	if x == nil {
		return 0
	}
	return len(x.errs)
}

// ErrorOrNil returns the error if non-empty, nil otherwise (inspired by go-multierror)
// Nil-safe: (*goerr.Errors)(nil).ErrorOrNil() returns nil
func (x *Errors) ErrorOrNil() error {
	if x == nil || len(x.errs) == 0 {
		return nil
	}
	return x
}

// Errors returns the slice of wrapped errors (inspired by go-multierror.WrappedErrors)
func (x *Errors) Errors() []error {
	if x == nil || len(x.errs) == 0 {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]error, len(x.errs))
	copy(result, x.errs)
	return result
}

// HasTag checks if any wrapped error has the specified tag
func (x *Errors) HasTag(tag tag) bool {
	if x == nil {
		return false
	}

	// Check wrapped errors
	for _, err := range x.errs {
		if HasTag(err, tag) {
			return true
		}
	}
	return false
}

// ErrorsJSON represents JSON structure for Errors
type ErrorsJSON struct {
	Errors []any `json:"errors"`
}

// MarshalJSON implements json.Marshaler interface for Errors type
func (x *Errors) MarshalJSON() ([]byte, error) {
	if x == nil {
		return []byte("null"), nil
	}

	result := ErrorsJSON{
		Errors: make([]any, len(x.errs)),
	}

	// Serialize each error
	for i, err := range x.errs {
		if goErr := Unwrap(err); goErr != nil {
			result.Errors[i] = goErr.Printable()
		} else if marshaler, ok := err.(json.Marshaler); ok {
			data, marshalErr := marshaler.MarshalJSON()
			if marshalErr != nil {
				result.Errors[i] = err.Error()
			} else {
				result.Errors[i] = json.RawMessage(data)
			}
		} else {
			result.Errors[i] = err.Error()
		}
	}

	return json.Marshal(result)
}

// LogValue implements slog.LogValuer interface for structured logging
func (x *Errors) LogValue() slog.Value {
	if x == nil {
		return slog.AnyValue(nil)
	}

	attrs := []slog.Attr{
		slog.Int("count", len(x.errs)),
	}

	// Add individual errors
	errorAttrs := make([]any, 0, len(x.errs)*2)
	for i, err := range x.errs {
		key := strconv.Itoa(i)
		if lv, ok := err.(slog.LogValuer); ok {
			errorAttrs = append(errorAttrs, key, lv.LogValue())
		} else {
			errorAttrs = append(errorAttrs, key, err.Error())
		}
	}
	attrs = append(attrs, slog.Group("errors", errorAttrs...))

	return slog.GroupValue(attrs...)
}

// Format implements fmt.Formatter interface
func (x *Errors) Format(s fmt.State, verb rune) {
	if x == nil {
		return
	}

	switch verb {
	case 'v':
		if s.Flag('+') {
			// Detailed format with all error details
			fmt.Fprintf(s, "Errors (%d):\n", len(x.errs))
			for i, err := range x.errs {
				fmt.Fprintf(s, "  [%d] %+v\n", i, err)
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
