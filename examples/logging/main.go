package main

import (
	"errors"
	"os"

	"log/slog"

	"github.com/m-mizutani/goerr"
)

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
		logger.Error("fail someAction", slog.Any("error", err))
	}
}
