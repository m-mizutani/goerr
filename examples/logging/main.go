package main

import (
	"os"

	"log/slog"

	"github.com/m-mizutani/goerr"
)

var runtimeError = goerr.New("runtime error")

func someAction(input string) error {
	if input != "OK" {
		return goerr.Wrap(runtimeError, "input is not OK").With("input", input)
	}
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := someAction("ng"); err != nil {
		logger.Error("fail someAction", slog.Any("error", err))
	}
}
