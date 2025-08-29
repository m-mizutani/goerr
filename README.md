# goerr [![test](https://github.com/m-mizutani/goerr/actions/workflows/test.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/test.yml) [![gosec](https://github.com/m-mizutani/goerr/actions/workflows/gosec.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/gosec.yml) [![package scan](https://github.com/m-mizutani/goerr/actions/workflows/trivy.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/trivy.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/m-mizutani/goerr.svg)](https://pkg.go.dev/github.com/m-mizutani/goerr)

Package `goerr` provides more contextual error handling in Go.

## Features

`goerr` provides the following features:

- Stack traces
  - Compatible with `github.com/pkg/errors`.
  - Structured stack traces with `goerr.Stack` is available.
- Contextual variables to errors using:
  - Key value data by `goerr.Value(key, value)` (or `goerr.V(key, value)` as alias).
  - **Type-safe key value data** by `goerr.TypedValue(key, value)` (or `goerr.TV(key, value)` as alias) with compile-time type checking.
  - Tag value data can be defined by `goerr.NewTag` and set into error by `goerr.Tag(tag)` (or `goerr.T(tag)` as alias).
- `errors.Is` to identify errors and `errors.As` to unwrap errors.
- `slog.LogValuer` interface to output structured logs with `slog`.

## Usage

You can install `goerr` by `go get`:

```sh
go get github.com/m-mizutani/goerr/v2
```

### Stack trace

`goerr` records stack trace when creating an error. The format is compatible with `github.com/pkg/errors` and it can be used for [sentry.io](https://sentry.io), etc.

```go
func someAction(fname string) error {
	if _, err := os.Open(fname); err != nil {
		return goerr.Wrap(err, "failed to open file")
	}
	return nil
}

func main() {
	if err := someAction("no_such_file.txt"); err != nil {
		log.Fatalf("%+v", err)
	}
}
```

Output:
```
2024/04/06 10:30:27 failed to open file: open no_such_file.txt: no such file or directory
main.someAction
        /Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/stacktrace_print/main.go:12
main.main
        /Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/stacktrace_print/main.go:18
runtime.main
        /usr/local/go/src/runtime/proc.go:271
runtime.goexit
        /usr/local/go/src/runtime/asm_arm64.s:1222
exit status 1
```

You can not only print the stack trace, but also extract the stack trace by `goerr.Unwrap(err).Stacks()`.

```go
if err := someAction("no_such_file.txt"); err != nil {
  // NOTE: `errors.Unwrap` also works
  if goErr := goerr.Unwrap(err); goErr != nil {
    for i, st := range goErr.Stacks() {
      log.Printf("%d: %v\n", i, st)
    }
  }
  log.Fatal(err)
}
```

`Stacks()` returns a slice of `goerr.Stack` struct, which contains `Func`, `File`, and `Line`.

```
2024/04/06 10:35:30 0: &{main.someAction /Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/stacktrace_extract/main.go 12}
2024/04/06 10:35:30 1: &{main.main /Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/stacktrace_extract/main.go 18}
2024/04/06 10:35:30 2: &{runtime.main /usr/local/go/src/runtime/proc.go 271}
2024/04/06 10:35:30 3: &{runtime.goexit /usr/local/go/src/runtime/asm_arm64.s 1222}
2024/04/06 10:35:30 failed to open file: open no_such_file.txt: no such file or directory
exit status 1
```

**NOTE**: If the error is wrapped by `goerr` multiply, `%+v` will print the stack trace of the deepest error.

**Tips**: If you want not to print the stack trace for current stack frame, you can use `Unstack` method. Also, `UnstackN` method removes the top multiple stack frames.

```go
if err := someAction("no_such_file.txt"); err != nil {
	// Unstack() removes the current stack frame from the error message.
	return goerr.Wrap(err, "failed to someAction").Unstack()
}
```

### Add/Extract contextual variables

`goerr` provides the `Value(key, value)` method to add contextual variables to errors. The standard way to handle errors in Go is by injecting values into error messages. However, this approach makes it difficult to aggregate various errors. On the other hand, `goerr`'s `Value` method allows for adding contextual information to errors without changing error message, making it easier to aggregate error logs. Additionally, error handling services like Sentry.io can handle errors more accurately with this feature.

```go
var errFormatMismatch = errors.New("format mismatch")

func someAction(tasks []task) error {
	for _, t := range tasks {
		if err := validateData(t.Data); err != nil {
			return goerr.Wrap(err, "failed to validate data", goerr.Value("name", t.Name))
		}
	}
	// ....
	return nil
}

func validateData(data string) error {
	if !strings.HasPrefix(data, "data:") {
		return goerr.Wrap(errFormatMismatch, goerr.Value("data", data))
	}
	return nil
}

type task struct {
	Name string
	Data string
}

func main() {
	tasks := []task{
		{Name: "task1", Data: "data:1"},
		{Name: "task2", Data: "invalid"},
		{Name: "task3", Data: "data:3"},
	}
	if err := someAction(tasks); err != nil {
		if goErr := goerr.Unwrap(err); goErr != nil {
			for k, v := range goErr.Values() {
				log.Printf("var: %s => %v\n", k, v)
			}
		}
		log.Fatalf("msg: %s", err)
	}
}
```

Output:
```
2024/04/06 14:40:59 var: data => invalid
2024/04/06 14:40:59 var: name => task2
2024/04/06 14:40:59 msg: failed to validate data: : format mismatch
exit status 1
```

If you want to send the error to sentry.io with [SDK](https://docs.sentry.io/platforms/go/), you can extract the contextual variables by `goErr.Values()` and set them to the scope.

```go
// Sending error to Sentry
hub := sentry.CurrentHub().Clone()
hub.ConfigureScope(func(scope *sentry.Scope) {
  if goErr := goerr.Unwrap(err); goErr != nil {
    for k, v := range goErr.Values() {
      scope.SetExtra(k, v)
    }
  }
})
evID := hub.CaptureException(err)
```

#### Type-safe values

`goerr` provides type-safe contextual variables using `TypedValue` with generics. This feature offers compile-time type checking, preventing type-related runtime errors and enabling better IDE support with autocompletion.

```go
// Define typed keys (usually at package level)
var (
    UserIDKey    = goerr.NewTypedKey[string]("user_id")
    RequestIDKey = goerr.NewTypedKey[int64]("request_id")
    ConfigKey    = goerr.NewTypedKey[*Config]("config")
)

type Config struct {
    Host string
    Port int
}

func handleRequest(userID string, requestID int64) error {
    config := &Config{Host: "localhost", Port: 8080}

    if err := validateUser(userID); err != nil {
        return goerr.Wrap(err, "user validation failed",
            goerr.TV(UserIDKey, userID),        // Type-safe: string
            goerr.TV(RequestIDKey, requestID),  // Type-safe: int64
            goerr.TV(ConfigKey, config),        // Type-safe: *Config
        )
    }
    return nil
}

func validateUser(userID string) error {
    if userID == "" {
        return goerr.New("invalid user ID",
            goerr.TV(UserIDKey, userID),
        )
    }
    return nil
}

func main() {
    if err := handleRequest("", 12345); err != nil {
        // Type-safe value retrieval (no type assertions needed)
        if userID, ok := goerr.GetTypedValue(err, UserIDKey); ok {
            log.Printf("Failed for user: %s", userID) // userID is string
        }

        if requestID, ok := goerr.GetTypedValue(err, RequestIDKey); ok {
            log.Printf("Request ID: %d", requestID) // requestID is int64
        }

        if config, ok := goerr.GetTypedValue(err, ConfigKey); ok {
            log.Printf("Config: %+v", config) // config is *Config
        }

        log.Fatal(err)
    }
}
```

**Benefits of type-safe values:**

- **Compile-time type checking**: Prevents type mismatches at compile time
- **IDE support**: Autocompletion for key names and value types
- **No type assertions**: Values are returned with their correct types
- **Backward compatibility**: Works alongside existing `Value()` method

**Error cases caught at compile time:**

```go
userKey := goerr.NewTypedKey[string]("user_id")

// ❌ This will NOT compile - type mismatch
// err := goerr.New("test", goerr.TV(userKey, 123))

// ❌ This will NOT compile - wrong key type for retrieval
// intKey := goerr.NewTypedKey[int]("user_id")
// value, ok := goerr.GetTypedValue(err, intKey)
```

#### Tags

There are use cases where we need to adjust the error handling strategy based on the nature of the error. A clear example is an HTTP server, where the status code to be returned varies depending on whether it's an error from a downstream system, a missing resource, or an unauthorized request. To handle this precisely, you could predefine errors for each type and use methods like `errors.Is` in the error handling section to verify and branch the processing accordingly. However, this approach becomes challenging as the program grows larger and the number and variety of errors increase.

`goerr` provides also `WithTags(tags ...string)` method to add tags to errors. Tags are useful when you want to categorize errors. For example, you can add tags like "critical" or "warning" to errors.

```go
var (
	ErrTagSysError   = goerr.NewTag("system_error")
	ErrTagBadRequest = goerr.NewTag("bad_request")
)

func handleError(w http.ResponseWriter, err error) {
	if goErr := goerr.Unwrap(err); goErr != nil {
		switch {
		case goErr.HasTag(ErrTagSysError):
			w.WriteHeader(http.StatusInternalServerError)
		case goErr.HasTag(ErrTagBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	_, _ = w.Write([]byte(err.Error()))
}

func someAction() error {
	if _, err := http.Get("http://example.com/some/resource"); err != nil {
		return goerr.Wrap(err, "failed to get some resource").WithTags(ErrTagSysError)
	}
	return nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := someAction(); err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	http.ListenAndServe(":8090", nil)
}
```

### Structured logging

`goerr` provides `slog.LogValuer` interface to output structured logs with `slog`. It can be used to output not only the error message but also the stack trace and contextual variables. Additionally, unwrapped errors can be output recursively.

```go
var errRuntime = errors.New("runtime error")

func someAction(input string) error {
	if err := validate(input); err != nil {
		return goerr.Wrap(err, "failed validation")
	}
	return nil
}

func validate(input string) error {
	if input != "OK" {
		return goerr.Wrap(errRuntime, "invalid input", goerr.V("input", input))
	}
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := someAction("ng"); err != nil {
		logger.Error("aborted myapp", slog.Any("error", err))
	}
}
```

Output:
```json
{
  "time": "2024-04-06T11:32:40.350873+09:00",
  "level": "ERROR",
  "msg": "aborted myapp",
  "error": {
    "message": "failed validation",
    "stacktrace": [
      "/Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/logging/main.go:16 main.someAction",
      "/Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/logging/main.go:30 main.main",
      "/usr/local/go/src/runtime/proc.go:271 runtime.main",
      "/usr/local/go/src/runtime/asm_arm64.s:1222 runtime.goexit"
    ],
    "cause": {
      "message": "invalid input",
      "values": {
        "input": "ng"
      },
      "stacktrace": [
        "/Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/logging/main.go:23 main.validate",
        "/Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/logging/main.go:15 main.someAction",
        "/Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/logging/main.go:30 main.main",
        "/usr/local/go/src/runtime/proc.go:271 runtime.main",
        "/usr/local/go/src/runtime/asm_arm64.s:1222 runtime.goexit"
      ],
      "cause": "runtime error"
    }
  }
}
```

### Builder

`goerr` provides `goerr.NewBuilder()` to create an error with pre-defined contextual variables. It is useful when you want to create an error with the same contextual variables in multiple places.

```go
type object struct {
	id    string
	color string
}

func (o *object) Validate() error {
	eb := goerr.NewBuilder(goerr.Value("id", o.id))

	if o.color == "" {
		return eb.New("color is empty")
	}

	return nil
}

func main() {
	obj := &object{id: "object-1"}

	if err := obj.Validate(); err != nil {
		slog.Default().Error("Validation error", "err", err)
	}
}
```

Output:
```
2024/10/19 14:19:54 ERROR Validation error err.message="color is empty" err.values.id=object-1 (snip)
```

## License

The 2-Clause BSD License. See [LICENSE](LICENSE) for more detail.
