package goerr

import (
	"fmt"
)

// Builder keeps a set of key-value pairs and can create a new error and wrap error with the key-value pairs.
type Builder struct {
	values values
}

// NewBuilder creates a new Builder
func NewBuilder() *Builder {
	return &Builder{values: make(values)}
}

// With copies the current Builder and adds a new key-value pair.
func (x *Builder) With(key string, value any) *Builder {
	newVS := &Builder{values: x.values.clone()}
	newVS.values[key] = value
	return newVS
}

// New creates a new error with message
func (x *Builder) New(format string, args ...any) *Error {
	err := newError()
	err.msg = fmt.Sprintf(format, args...)
	err.values = x.values.clone()

	return err
}

// Wrap creates a new Error with caused error and add message.
func (x *Builder) Wrap(cause error, msg ...any) *Error {
	err := newError()
	err.msg = toWrapMessage(msg)
	err.cause = cause
	err.values = x.values.clone()
	return err
}
