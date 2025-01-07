package goerr

// Builder keeps a set of key-value pairs and can create a new error and wrap error with the key-value pairs.
type Builder struct {
	options []Option
}

// NewBuilder creates a new Builder
func NewBuilder(options ...Option) *Builder {
	return &Builder{
		options: options,
	}
}

// With copies the current Builder and adds a new key-value pair.
//
// Deprecated: Use goerr.Value instead.
func (x *Builder) With(key string, value any) *Builder {
	newBuilder := &Builder{
		options: x.options[:],
	}
	newBuilder.options = append(newBuilder.options, Value(key, value))
	return newBuilder
}

// New creates a new error with message
func (x *Builder) New(msg string, options ...Option) *Error {
	err := newError(append(x.options, options...)...)
	err.msg = msg
	return err
}

// Wrap creates a new Error with caused error and add message.
func (x *Builder) Wrap(cause error, msg string, options ...Option) *Error {
	err := newError(append(x.options, options...)...)
	err.msg = msg
	err.cause = cause
	return err
}
