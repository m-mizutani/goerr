# goerr [![test](https://github.com/m-mizutani/goerr/actions/workflows/test.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/test.yml) [![gosec](https://github.com/m-mizutani/goerr/actions/workflows/gosec.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/gosec.yml) [![package scan](https://github.com/m-mizutani/goerr/actions/workflows/trivy.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/trivy.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/m-mizutani/goerr.svg)](https://pkg.go.dev/github.com/m-mizutani/goerr)

Package `goerr` provides more contextual error handling in Go.

## Features

`goerr` provides the following features:

- Stack traces
  - Compatible with `github.com/pkg/errors`.
  - Structured stack traces with `goerr.Stack` is available.
- Contextual variables to errors using `With(key, value)`.
- `errors.Is` to identify errors and `errors.As` to unwrap errors.
- `slog.LogValuer` interface to output structured logs with `slog`.

## Usage

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


`goerr` provides the `With(key, value)` method to add contextual variables to errors. The standard way to handle errors in Go is by injecting values into error messages. However, this approach makes it difficult to aggregate various errors. On the other hand, `goerr`'s `With` method allows for adding contextual information to errors without changing error message, making it easier to aggregate error logs. Additionally, error handling services like Sentry.io can handle errors more accurately with this feature.

```go
var errFormatMismatch = errors.New("format mismatch")

func someAction(tasks []task) error {
	for _, t := range tasks {
		if err := validateData(t.Data); err != nil {
			return goerr.Wrap(err, "failed to validate data").With("name", t.Name)
		}
	}
	// ....
	return nil
}

func validateData(data string) error {
	if !strings.HasPrefix(data, "data:") {
		return goerr.Wrap(errFormatMismatch).With("data", data)
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
		return goerr.Wrap(errRuntime, "invalid input").With("input", input)
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
	eb := goerr.NewBuilder().With("id", o.id)

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
