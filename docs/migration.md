# Migration Guide

This guide helps you migrate to goerr v2 from other error handling libraries.

## Table of Contents

- [From github.com/pkg/errors](#from-githubcompkgerrors)
- [From standard library errors](#from-standard-library-errors)
- [From goerr v1 to v2](#from-goerr-v1-to-v2)

## From github.com/pkg/errors

goerr is designed to be compatible with `github.com/pkg/errors`, making migration straightforward.

### Stack Trace Compatibility

goerr's stack traces are fully compatible with pkg/errors format:

```go
// pkg/errors
import "github.com/pkg/errors"
err := errors.New("something failed")
err = errors.Wrap(err, "operation failed")
fmt.Printf("%+v", err)  // Prints with stack trace

// goerr - same behavior
import "github.com/m-mizutani/goerr/v2"
err := goerr.New("something failed")
err = goerr.Wrap(err, "operation failed")
fmt.Printf("%+v", err)  // Prints with stack trace (same format)
```

### Migration Table

| pkg/errors | goerr | Notes |
|------------|-------|-------|
| `errors.New()` | `goerr.New()` | Direct replacement |
| `errors.Wrap()` | `goerr.Wrap()` | Direct replacement |
| `errors.Wrapf()` | `goerr.Wrap(err, fmt.Sprintf(...))` | Use fmt.Sprintf for formatting |
| `errors.WithStack()` | `goerr.Wrap(err, err.Error())` | Wraps with same message |
| `errors.Cause()` | See below* | Requires loop for root cause |
| `errors.Is()` | `errors.Is()` (standard) | Use standard library |
| `errors.As()` | `errors.As()` (standard) | Use standard library |

*For `errors.Cause()` replacement:
```go
// pkg/errors.Cause() gets the root cause
rootCause := errors.Cause(err)

// goerr equivalent - loop to get root cause
func getRootCause(err error) error {
    for {
        unwrapped := errors.Unwrap(err)
        if unwrapped == nil {
            return err
        }
        err = unwrapped
    }
}
rootCause := getRootCause(err)
```

### Additional Benefits

When migrating to goerr, you gain additional features:

```go
// Add contextual variables
err := goerr.Wrap(originalErr, "failed to process",
    goerr.Value("user_id", userID),
    goerr.Value("retry_count", retries))

// Add error tags for categorization
err = goerr.Wrap(err, "database error",
    goerr.Tag(ErrTagDatabase))

// Use with structured logging
logger.Error("operation failed", slog.Any("error", err))
```

## From standard library errors

Migrating from standard `errors` package to goerr adds rich context while maintaining compatibility.

### Basic Migration

```go
// Standard errors
import "errors"
import "fmt"

var ErrNotFound = errors.New("not found")

func process() error {
    if err := doWork(); err != nil {
        return fmt.Errorf("process failed: %w", err)
    }
    return nil
}

// goerr equivalent
import "github.com/m-mizutani/goerr/v2"

var ErrNotFound = goerr.New("not found", goerr.ID("ERR_NOT_FOUND"))

func process() error {
    if err := doWork(); err != nil {
        return goerr.Wrap(err, "process failed")
    }
    return nil
}
```

### Error Comparison

```go
// Standard errors - sentinel errors
var ErrInvalid = errors.New("invalid input")

if errors.Is(err, ErrInvalid) {
    // handle invalid input
}

// goerr - with ID for flexible comparison
var ErrInvalid = goerr.New("invalid input", goerr.ID("ERR_INVALID"))

if errors.Is(err, ErrInvalid) {
    // handle invalid input - matches by ID
}
```

### Adding Context (Major Benefit)

Standard errors lose context during wrapping. goerr preserves and enriches it:

```go
// Standard errors - context in message only
return fmt.Errorf("failed to save user %s: %w", userID, err)

// goerr - structured context
return goerr.Wrap(err, "failed to save user",
    goerr.Value("user_id", userID),
    goerr.Value("timestamp", time.Now()))

// Later, extract the context
if goErr := goerr.Unwrap(err); goErr != nil {
    userID := goErr.Values()["user_id"]
    // Use the structured data
}
```

## From goerr v1 to v2

### Import Path Change

```go
// v1
import "github.com/m-mizutani/goerr"

// v2
import "github.com/m-mizutani/goerr/v2"
```

### Major Changes in v2

#### 1. Type-Safe Values (New Feature)

v2 introduces compile-time type checking for contextual values:

```go
// v1 - only string-based values
err := goerr.New("error").With("user_id", userID)

// v2 - type-safe option available
var UserIDKey = goerr.NewTypedKey[string]("user_id")
err := goerr.New("error", goerr.TV(UserIDKey, userID))

// Type-safe retrieval
if userID, ok := goerr.GetTypedValue(err, UserIDKey); ok {
    // userID is guaranteed to be string type
}
```

#### 2. Options Pattern

v2 uses functional options pattern consistently:

```go
// v1
err := goerr.New("error").With("key", value).WithTags("tag1")

// v2
err := goerr.New("error",
    goerr.Value("key", value),
    goerr.Tag(TagValue))
```

#### 3. Error ID Support

v2 adds error IDs for flexible comparison:

```go
// v2 only
var ErrDatabase = goerr.New("database error", goerr.ID("ERR_DB"))

// Errors with same ID are considered equal
if errors.Is(err, ErrDatabase) {
    // Matches any error with ID "ERR_DB"
}
```

#### 4. Builder Pattern Enhanced

```go
// v1
err := goerr.New("error").With("common", value)

// v2 - Builder for shared context
builder := goerr.NewBuilder(
    goerr.Value("service", "api"),
    goerr.Value("version", "2.0"))

err1 := builder.New("error 1")
err2 := builder.Wrap(otherErr, "error 2")
```

### Backward Compatibility

Most v1 code works with minimal changes:
- Core functions (`New`, `Wrap`, `Unwrap`) remain the same
- Stack trace format unchanged
- Standard error interface compatibility maintained

### Migration Checklist

- [ ] Update import path to `/v2`
- [ ] Convert `.With()` chains to option parameters
- [ ] Replace `.WithTags()` with `goerr.Tag()` options
- [ ] Consider using typed keys for critical values
- [ ] Add error IDs for sentinel errors
- [ ] Update error assertions if using typed values

## Best Practices After Migration

### 1. Define Package-Level Keys and Tags

```go
package myapp

import "github.com/m-mizutani/goerr/v2"

// Typed keys for common values
var (
    UserIDKey    = goerr.NewTypedKey[string]("user_id")
    RequestIDKey = goerr.NewTypedKey[string]("request_id")
)

// Error tags for categorization
var (
    TagValidation = goerr.NewTag("validation")
    TagDatabase   = goerr.NewTag("database")
    TagExternal   = goerr.NewTag("external")
)

// Sentinel errors with IDs
var (
    ErrNotFound = goerr.New("not found", goerr.ID("ERR_NOT_FOUND"))
    ErrInvalid  = goerr.New("invalid", goerr.ID("ERR_INVALID"))
)
```

### 2. Use Structured Logging

```go
logger := slog.Default()

if err := doWork(); err != nil {
    logger.Error("work failed", slog.Any("error", err))
    // Automatically includes stack trace, values, and tags
}
```

### 3. Consistent Error Handling

```go
func handleError(w http.ResponseWriter, err error) {
    if goErr := goerr.Unwrap(err); goErr != nil {
        // Log with full context
        logger.Error("request failed", slog.Any("error", err))
        
        // Respond based on tags
        switch {
        case goErr.HasTag(TagValidation):
            w.WriteHeader(http.StatusBadRequest)
        case goErr.HasTag(TagDatabase):
            w.WriteHeader(http.StatusServiceUnavailable)
        default:
            w.WriteHeader(http.StatusInternalServerError)
        }
    }
}
```

## Need Help?

- Check the [examples](../examples) directory for working code samples
- Refer to the [API documentation](https://pkg.go.dev/github.com/m-mizutani/goerr/v2)
- File issues on [GitHub](https://github.com/m-mizutani/goerr/issues)