# goerr [![Build Status](https://travis-ci.com/m-mizutani/goerr.svg?branch=master)](https://travis-ci.com/m-mizutani/goerr) [![Go Reference](https://pkg.go.dev/badge/github.com/m-mizutani/goerr.svg)](https://pkg.go.dev/github.com/m-mizutani/goerr)

Package `goerr` provides powerful error handling features in Go.

- Stacktrace (Compatible with `github.com/pkg/errors`)
- Keep variables related to error by `With()`

## Usage

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
		return goerr.New("input is not OK").With("input", input)
	}
	return nil
}

func main() {
	if err := someAction("ng"); err != nil {
		var goErr *goerr.Error
		if errors.As(err, &goErr) {
			log.Printf("Values: %+v\n", goErr.Values())
		}
		log.Fatalf("Error: %+v\n", err)
	}
}

```

Output:
```
2021/04/10 09:29:10 Values: map[input:ng]
2021/04/10 09:29:10 Error: input is not OK
main.someAction
	/xxx/github.com/m-mizutani/goerr/examples/basic/main.go:12
main.main
	/xxx/github.com/m-mizutani/goerr/examples/basic/main.go:18
runtime.main
	/yyy/src/runtime/proc.go:204
runtime.goexit
	/yyy/src/runtime/asm_amd64.s:1374
exit status 1
```

## License

The 2-Clause BSD License. See [LICENSE](LICENSE) for more detail.
