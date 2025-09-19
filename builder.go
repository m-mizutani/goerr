package goerr

// Builder keeps a set of key-value pairs and can create a new error and wrap error with the key-value pairs.
type Builder struct {
	options []Option
}

// NewBuilder creates a new Builder.
// A Builder is useful for creating multiple errors that share a common context, such as a request ID or service name, without repeatedly specifying the same options.
//
// Usage:
//   builder := goerr.NewBuilder(goerr.V("service", "auth"), goerr.V("request_id", "req123"))
//   err1 := builder.New("user not found")
//   err2 := builder.Wrap(dbErr, "query failed")
//   // Both errors will include service and request_id context
func NewBuilder(options ...Option) *Builder {
	return &Builder{
		options: options,
	}
}

// With copies the current Builder and adds a new key-value pair.
//
// Usage:
//   baseBuilder := goerr.NewBuilder(goerr.V("service", "auth"))
//   userBuilder := baseBuilder.With(goerr.V("user_id", "user123"))
//   err := userBuilder.New("access denied") // includes both service and user_id
func (x *Builder) With(options ...Option) *Builder {
	newBuilder := &Builder{
		options: x.options[:],
	}
	newBuilder.options = append(newBuilder.options, options...)
	return newBuilder
}

// New creates a new error with message
//
// Usage:
//   builder := goerr.NewBuilder(goerr.V("service", "auth"))
//   err := builder.New("authentication failed") // includes service context
func (x *Builder) New(msg string, options ...Option) *Error {
	err := newError(append(x.options, options...)...)
	err.msg = msg
	return err
}

// Wrap creates a new Error with caused error and add message.
//
// Usage:
//   builder := goerr.NewBuilder(goerr.V("service", "auth"))
//   err := builder.Wrap(dbErr, "database query failed") // wraps dbErr with service context
func (x *Builder) Wrap(cause error, msg string, options ...Option) *Error {
	err := newError(append(x.options, options...)...)
	err.msg = msg
	err.cause = cause
	return err
}
