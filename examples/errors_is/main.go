package main

import (
	"errors"
	"log"

	"github.com/abyssparanoia/goerr"
)

var errInvalidInput = errors.New("invalid input")

func someAction(input string) error {
	if input != "OK" {
		return goerr.Wrap(errInvalidInput, "input is not OK").WithValue("input", input)
	}
	// .....
	return nil
}

func main() {
	if err := someAction("ng"); err != nil {
		switch {
		case errors.Is(err, errInvalidInput):
			log.Printf("It's user's bad: %v\n", err)
		}
	}
}
