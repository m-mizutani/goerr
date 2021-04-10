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
