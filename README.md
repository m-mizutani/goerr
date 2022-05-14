# goerr [![test](https://github.com/m-mizutani/goerr/actions/workflows/test.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/test.yml) [![gosec](https://github.com/m-mizutani/goerr/actions/workflows/gosec.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/gosec.yml) [![package scan](https://github.com/m-mizutani/goerr/actions/workflows/trivy.yml/badge.svg)](https://github.com/m-mizutani/goerr/actions/workflows/trivy.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/m-mizutani/goerr.svg)](https://pkg.go.dev/github.com/m-mizutani/goerr)

Package `goerr` provides more contextual error handling in Go.

## Features

- Adding contextual variables to error by `With(key, value)`
- Records stacktrace (Compatible with `github.com/pkg/errors`)
- Supports `errors.Is` to identify error and `errors.As` to unwrap error
- Provides structured stacktrace and contextual variables

## Usage

### Extract values

Example code is [here](examples/basic/main.go)
```go
package main

import (
	"errors"
	"log"

	"github.com/m-mizutani/goerr"
)

func someAction(input string) error {
	if input != "OK" {
		return goerr.New("input is not OK").With("input", input).With("time", time.Now())
	}
	return nil
}

func main() {
	if err := someAction("ng"); err != nil {
		var goErr *goerr.Error
		if errors.As(err, &goErr) {
			for k, v := range goErr.Values() {
				log.Printf("%s = %v\n", k, v)
			}
		}
		log.Fatalf("Error: %+v\n", err)
	}
}
```

Output:
```
2022/05/14 10:28:08 input = ng
2022/05/14 10:28:08 time = 2022-05-14 10:28:08.452831 +0900 JST m=+0.000483668
2022/05/14 10:28:08 Error: input is not OK
main.someAction
        /xxx/goerr/examples/basic/main.go:13
main.main
        /xxx/goerr/examples/basic/main.go:19
runtime.main
        /usr/local/go/src/runtime/proc.go:250
runtime.goexit
        /usr/local/go/src/runtime/asm_arm64.s:1259
exit status 1
```

### Extract stack trace

```go
import (
	"github.com/m-mizutani/goerr"

	"github.com/rs/zerolog/log"
)

func someAction(input string) error {
	if input != "OK" {
		return goerr.New("input is not OK").With("input", input)
	}
	return nil
}

func main() {
	if err := someAction("ng"); err != nil {
		// Same with errors.As extraction
		if goErr := goerr.Unwrap(err); goErr != nil {
			stacks := goErr.Stacks()
			log.Info().Interface("stackTrace", stacks).Msg("Show stacktrace")
		}
	}
}
```

Output:
```json
{
  "level": "info",
  "stackTrace": [
    {
      "func": "main.someAction",
      "file": "/Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/stacktrace/main.go",
      "line": 11
    },
    {
      "func": "main.main",
      "file": "/Users/mizutani/.ghq/github.com/m-mizutani/goerr/examples/stacktrace/main.go",
      "line": 17
    },
    {
      "func": "runtime.main",
      "file": "/usr/local/go/src/runtime/proc.go",
      "line": 250
    },
    {
      "func": "runtime.goexit",
      "file": "/usr/local/go/src/runtime/asm_arm64.s",
      "line": 1259
    }
  ],
  "time": "2022-05-14T10:50:42+09:00",
  "message": "Show stacktrace"
}
```


## License

The 2-Clause BSD License. See [LICENSE](LICENSE) for more detail.
