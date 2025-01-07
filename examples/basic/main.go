package main

import (
	"errors"
	"log"
	"time"

	"github.com/m-mizutani/goerr/v2"
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
