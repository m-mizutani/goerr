package goerr

// Builder keeps a set of key-value pairs and can create a new error and wrap error with the key-value pairs.
type Builder struct {
	options []Option
}

// NewBuilder creates a new Builder.
// A Builder is useful for creating multiple errors that share a common context, such as a request ID or service name, without repeatedly specifying the same options.
func NewBuilder(options ...Option) *Builder {
	return &Builder{
		options: options,
	}
}

// With copies the current Builder and adds a new key-value pair.
func (x *Builder) With(options ...Option) *Builder {
	newBuilder := &Builder{
		options: x.options[:],
	}
	newBuilder.options = append(newBuilder.options, options...)
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
