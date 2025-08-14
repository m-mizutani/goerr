# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Restriction & Rules for Development

- In principle, do not trust developers who use this library from outside
  - Do not export unnecessary methods, structs, and variables
  - Assume that exposed items will be changed. Never expose fields that would be problematic if changed
  - Use `export_test.go` for items that need to be exposed for testing purposes
- When making changes, before finishing the task, always:
  - Run `go vet ./...`, `go fmt ./...` to format the code
  - Run `golangci-lint run ./...` to check lint error
  - Run `gosec -quiet -exclude-generated ./...` to check security issue
  - Run ALL tests by `go test ./...` to ensure no impact on other code
- All comments and character literals in source code must be in English
- Test files should have `package {name}_test`. Do not use same package name
- A name of test file MUST be `xxx_test.go` for `xxx.go`. DO NOT allow out of the rule, e.g. `xxx_integration_test.go`, `xxx_e2e_test.go`, etc.
- Use named empty structure (e.g. `type ctxHogeKey struct{}` ) as private context key
- DO NOT get environment variables from GetEnv and LookupEnv in application source code (It's allowed in test code). You MUST obtain environment variables via only "github.com/urfave/cli/v3"
- If you need to create a test code file as `main` package, create it into `./tmp/playground`
- Use `go run .` instead of creating binary

## Development Commands

### Testing
- `go test .` - Run all tests in the current package
- `go vet .` - Run static analysis

### Build
- `go build` - Build the package
- `go mod tidy` - Clean up dependencies

## Architecture

This is a Go error handling library that provides enhanced error capabilities including stack traces, contextual variables, tags, and structured logging support.

### Core Components

**Error Type (`errors.go:106-113`)**
- `Error` struct is the main error type containing message, ID, stack trace, cause, values, and tags
- Implements standard Go error interface plus additional methods for contextual data
- Uses UUID for error identification and supports error wrapping chains

**Stack Traces (`stack.go`)**
- Compatible with `github.com/pkg/errors` format
- `Stack` struct represents individual stack frames with function, file, and line info
- `StackTrace` type provides formatting compatible with pkg/errors
- Stack collection starts at runtime.Callers depth 4 to skip library internals

**Tags (`tag.go`)**
- Tag system for error categorization using unexported `tag` struct
- Tags are created via `NewTag(value)` and applied via `Tag(t)` option
- Support for checking tag presence with `HasTag()` method
- Enables error classification for different handling strategies

**Builder Pattern (`builder.go`)**
- `Builder` struct allows pre-configuration of error options
- Supports creating multiple errors with shared contextual data
- Methods: `With()` for option chaining, `New()` and `Wrap()` for error creation

**Options System**
- Functional options pattern for error configuration
- `Value(key, value)` / `V()` for key-value pairs
- `Tag(tag)` / `T()` for error tags
- Options are applied during error creation and can be merged

### Key Features

**Contextual Variables**
- Errors can carry arbitrary key-value data via `Value()` option
- Values are merged across wrapped error chains
- Accessible via `Values()` method returning `map[string]any`

**JSON Serialization**
- `Printable()` method returns JSON-serializable struct
- Includes message, ID, stack trace, cause, values, and tags
- Direct JSON marshaling of `Error` type shows empty fields (private fields not serialized)

**Structured Logging**
- Implements `slog.LogValuer` interface for structured logging
- Outputs comprehensive error information including nested causes
- Stack traces formatted as strings with file:line function format

**Error Wrapping & Unwrapping**
- Standard `Unwrap()` method for error chain traversal
- `goerr.Unwrap()` function extracts goerr.Error from any error
- Maintains cause chain while adding contextual information

**Compatibility**
- Standard Go error interface compliance
- `errors.Is()` and `errors.As()` support
- Stack trace format compatible with github.com/pkg/errors
