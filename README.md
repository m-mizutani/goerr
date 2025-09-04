# goerr [![test](https://github.com/m-mizutani/goerr/actions/workflows/test.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/test.yml) [![gosec](https://github.com/m-mizutani/goerr/actions/workflows/gosec.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/gosec.yml) [![package scan](https://github.com/m-mizutani/goerr/actions/workflows/trivy.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/trivy.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/m-mizutani/goerr.svg)](https://pkg.go.dev/github.com/m-mizutani/goerr)

Enhanced error handling for Go with stack traces, contextual values, and structured logging

## Overview

`goerr` is a powerful error handling library for Go that enhances errors with rich contextual information. It provides stack traces, contextual variables, error categorization, and seamless integration with structured logging - all while maintaining full compatibility with Go's standard error handling patterns.

## Key Features

- **Rich Stack Traces**: Automatic capture with `github.com/pkg/errors` compatibility
- **Contextual Variables**: Attach key-value pairs to errors for better debugging
- **Type-Safe Values**: Compile-time type checking for error context using generics
- **Multiple Error Handling**: Aggregate and manage multiple errors with `goerr.Errors`
- **Error Categorization**: Tag-based error classification for different handling strategies
- **Structured Logging**: Native `slog` integration with recursive error details
- **Error Identity**: ID-based error comparison for flexible error matching
- **Builder Pattern**: Efficient error creation with pre-configured context
- **JSON Serialization**: Full error details in JSON format for APIs and logging

## Installation

```sh
go get github.com/m-mizutani/goerr/v2
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/m-mizutani/goerr/v2"
)

func main() {
    if err := processFile("data.txt"); err != nil {
        // Print error with stack trace
        log.Fatalf("%+v", err)
    }
}

func processFile(filename string) error {
    _, err := readFile(filename)
    if err != nil {
        return goerr.Wrap(err, "failed to process file",
            goerr.Value("filename", filename))
    }
    return nil
}

func readFile(filename string) error {
    // Simulate error
    return goerr.New("file not found")
}
```

## Core Features

### Error Creation and Wrapping

Create new errors or wrap existing ones with additional context:

```go
// Create a new error
err := goerr.New("validation failed")

// Wrap an existing error
if err := someFunc(); err != nil {
    return goerr.Wrap(err, "operation failed")
}

// Add contextual information
err = goerr.Wrap(err, "user not found",
    goerr.Value("user_id", userID),
    goerr.Value("timestamp", time.Now()))

// Extract goerr.Error from any error
if goErr := goerr.Unwrap(err); goErr != nil {
    values := goErr.Values() // Get all contextual values
}
```

### Multiple Error Handling

Aggregate multiple errors with `goerr.Errors`:

```go
// Collect errors during processing
var errs *goerr.Errors
for _, item := range items {
    if err := processItem(item); err != nil {
        errs = goerr.Append(errs, err)  // nil-safe
    }
}

// Return only if errors occurred
return errs.ErrorOrNil()  // nil if no errors

// Join errors directly
combined := goerr.Join(err1, err2, err3)

// All errors displayed together
fmt.Printf("%v", combined)
// Output: error1\nerror2\nerror3

// Works with standard library
if errors.Is(combined, err1) { /* true */ }
```

### Contextual Data

**String-based Values**

Attach arbitrary key-value pairs to errors:

```go
func validateUser(userID string, age int) error {
    if age < 18 {
        return goerr.New("user too young",
            goerr.V("user_id", userID),  // V is alias for Value
            goerr.V("age", age),
            goerr.V("required_age", 18))
    }
    return nil
}

// Extract values from error
if err := validateUser("user123", 16); err != nil {
    if goErr := goerr.Unwrap(err); goErr != nil {
        for key, value := range goErr.Values() {
            log.Printf("%s: %v", key, value)
        }
    }
}
```

**Type-safe Values**

Use compile-time type checking for error context:

```go
// Define typed keys (typically at package level)
var (
    UserIDKey    = goerr.NewTypedKey[string]("user_id")
    RequestIDKey = goerr.NewTypedKey[int64]("request_id")
    ConfigKey    = goerr.NewTypedKey[*Config]("config")
)

// Use typed values - compile-time type checking
err := goerr.New("validation failed",
    goerr.TV(UserIDKey, "user123"),      // Must be string
    goerr.TV(RequestIDKey, int64(42)),   // Must be int64
    goerr.TV(ConfigKey, currentConfig))  // Must be *Config

// Retrieve typed values - no type assertion needed
if userID, ok := goerr.GetTypedValue(err, UserIDKey); ok {
    // userID is string type, guaranteed
    fmt.Printf("User: %s\n", userID)
}
```

**Error Tags**

Categorize errors for different handling strategies:

```go
// Define tags
var (
    ErrTagNotFound   = goerr.NewTag("not_found")
    ErrTagValidation = goerr.NewTag("validation")
    ErrTagExternal   = goerr.NewTag("external")
)

// Tag errors
if user == nil {
    return goerr.New("user not found",
        goerr.T(ErrTagNotFound))  // T is alias for Tag
}

// Handle errors based on tags
if goerr.HasTag(err, ErrTagNotFound) {
    w.WriteHeader(http.StatusNotFound)
} else if goerr.HasTag(err, ErrTagValidation) {
    w.WriteHeader(http.StatusBadRequest)
} else {
    w.WriteHeader(http.StatusInternalServerError)
}
```

### Stack Traces

Stack traces are automatically captured and compatible with `github.com/pkg/errors`:

```go
func doWork() error {
    return goerr.New("something went wrong")
}

func main() {
    if err := doWork(); err != nil {
        // Print with stack trace using %+v
        log.Printf("%+v", err)
        
        // Extract stack programmatically
        if goErr := goerr.Unwrap(err); goErr != nil {
            for _, frame := range goErr.Stacks() {
                log.Printf("  at %s:%d in %s", 
                    frame.File, frame.Line, frame.Func)
            }
        }
    }
}

// Remove current frame from stack (useful for helper functions)
func helperFunc() error {
    return goerr.New("error from helper").Unstack()
}
```

## Advanced Features

### Error Identification

Use IDs for flexible error comparison:

```go
var (
    ErrInvalidInput = goerr.New("invalid input", goerr.ID("ERR_INVALID_INPUT"))
    ErrTimeout      = goerr.New("operation timeout", goerr.ID("ERR_TIMEOUT"))
)

func process() error {
    return goerr.Wrap(ErrInvalidInput, "validation failed",
        goerr.Value("field", "email"))
}

// Check error identity
if err := process(); err != nil {
    if errors.Is(err, ErrInvalidInput) {
        // Matches by ID, not pointer
        handleValidationError(err)
    }
}
```

### Builder Pattern

Create multiple errors with shared context:

```go
type Service struct {
    userID string
    reqID  string
}

func (s *Service) process() error {
    // Create builder with common context
    eb := goerr.NewBuilder(
        goerr.Value("user_id", s.userID),
        goerr.Value("request_id", s.reqID))
    
    // Use builder for multiple errors
    if err := s.validate(); err != nil {
        return eb.Wrap(err, "validation failed")
    }
    
    if err := s.save(); err != nil {
        return eb.Wrap(err, "save failed")
    }
    
    return nil
}
```

### Structured Logging

Native integration with Go's `slog` package:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

err := goerr.New("database error",
    goerr.Value("table", "users"),
    goerr.Value("operation", "insert"))

// Error implements slog.LogValuer
logger.Error("operation failed", slog.Any("error", err))

// Output (formatted):
// {
//   "level": "ERROR",
//   "msg": "operation failed",
//   "error": {
//     "message": "database error",
//     "values": {"table": "users", "operation": "insert"},
//     "stacktrace": [...]
//   }
// }
```

### JSON Serialization

Export full error details as JSON:

```go
err := goerr.New("validation error",
    goerr.Value("field", "email"),
    goerr.Tag(ValidationTag))

// Get JSON-serializable struct
printable := goerr.Unwrap(err).Printable()

// Or marshal directly
jsonData, _ := json.Marshal(err)

// Output includes message, stack trace, values, tags, and cause chain
```

## Examples

See the [examples](./examples) directory for complete working examples:
- Stack trace handling
- Contextual variables
- Multiple error aggregation
- HTTP error responses
- Sentry integration
- Structured logging with slog
- And more...

## Migration Guide

See [Migration Guide](./docs/migration.md) for migrating from:
- `github.com/pkg/errors`
- Standard library `errors` package
- goerr v1 to v2

## License

The 2-Clause BSD License. See [LICENSE](LICENSE) for more detail.