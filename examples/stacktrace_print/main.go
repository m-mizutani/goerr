package main

import (
	"errors"
	"log"

	"github.com/m-mizutani/goerr"
)

func nestedAction2() error {
	return errors.New("fatal error in the nested action2")
}

func nestedAction() error {
	return goerr.Wrap(nestedAction2(), "nestedAction2 failed")
}

func someAction() error {
	return goerr.Wrap(nestedAction(), "nestedAction failed")
}

func main() {
	if err := someAction(); err != nil {
		log.Fatalf("%+v", err)
	}
}
