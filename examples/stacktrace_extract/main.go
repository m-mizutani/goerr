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
		/*
			// NOTE: errors.As also works
			var goErr *goerr.Error
			if errors.As(err, &goErr); goErr != nil {
		*/
		if goErr := goerr.Unwrap(err); goErr != nil {
			for i, st := range goErr.Stacks() {
				log.Printf("%d: %+v\n", i, st)
			}
		}
		log.Fatal(err)
	}
}
