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
		log.Fatalf("%+v", err)
	}
}
