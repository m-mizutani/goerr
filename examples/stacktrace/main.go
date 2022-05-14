package main

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
		if goErr := goerr.Unwrap(err); goErr != nil {
			stacks := goErr.Stacks()
			log.Info().Interface("stackTrace", stacks).Msg("Show stacktrace")
		}
	}
}
