package main

import (
	"log"

	"github.com/m-mizutani/goerr"
)

func nestedAction2() error {
	return goerr.New("fatal error in the nested action2")
}

func nestedAction() error {
	return goerr.Wrap(nestedAction2(), "nestedAction2 failed")
}

func someAction() error {
	if err := nestedAction(); err != nil {
		return goerr.Wrap(err, "nestedAction failed")
	}
	return nil
}

func main() {
	if err := someAction(); err != nil {
		log.Fatalf("%+v", err)
	}
}
