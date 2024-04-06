package main

import (
	"log"
	"os"

	"github.com/m-mizutani/goerr"
)

func someAction(fname string) error {
	if _, err := os.Open(fname); err != nil {
		return goerr.Wrap(err, "failed to open file")
	}
	return nil
}

func main() {
	if err := someAction("no_such_file.txt"); err != nil {
		// NOTE: `errors.Unwrap` also works
		if goErr := goerr.Unwrap(err); goErr != nil {
			for i, st := range goErr.Stacks() {
				log.Printf("%d: %v\n", i, st)
			}
		}
		log.Fatal(err)
	}
}
