package goerr

// TypedKey represents a type-safe key for error values
type TypedKey[T any] struct {
	name string
}

// NewTypedKey creates a new type-safe key with the given name.
// This key can then be used with TV() and GetTypedValue() to attach and retrieve strongly-typed values from an error, providing compile-time safety.
func NewTypedKey[T any](name string) TypedKey[T] {
	return TypedKey[T]{name: name}
}

// String returns the string representation of the key for debugging
func (k TypedKey[T]) String() string {
	return k.name
}

// Name returns the name of the key
func (k TypedKey[T]) Name() string {
	return k.name
}

// TypedValue sets typed key and value to the error
func TypedValue[T any](key TypedKey[T], value T) Option {
	return func(err *Error) {
		err.typedValues[key.name] = value
	}
}

// TV is alias of TypedValue
func TV[T any](key TypedKey[T], value T) Option {
	return TypedValue(key, value)
}

// TypedValues returns map of key and value that is set by TypedValue. All wrapped goerr.Error typed key and values will be merged. Key and values of wrapped error is overwritten by upper goerr.Error.
func TypedValues(err error) map[string]any {
	if e := Unwrap(err); e != nil {
		return e.TypedValues()
	}

	return nil
}

// GetTypedValue returns value associated with the typed key from the error. It searches through the error chain.
func GetTypedValue[T any](err error, key TypedKey[T]) (T, bool) {
	if e := Unwrap(err); e != nil {
		return getTypedValueFromError(e, key)
	}

	var zero T
	return zero, false
}

func getTypedValueFromError[T any](err *Error, key TypedKey[T]) (T, bool) {
	// Search in current error's typed values
	if value, ok := err.typedValues[key.name]; ok {
		// Key found at this level. This is the definitive value.
		// Check if the type matches.
		if typedValue, ok := value.(T); ok {
			return typedValue, true
		}
		// Type does not match. Do not search deeper for this key.
		var zero T
		return zero, false
	}

	// Key not found at this level. Search in wrapped errors recursively.
	if cause := err.Unwrap(); cause != nil {
		if wrappedErr := Unwrap(cause); wrappedErr != nil {
			return getTypedValueFromError(wrappedErr, key)
		}
	}

	var zero T
	return zero, false
}
