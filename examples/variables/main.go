package main

import (
	"errors"
	"log"
	"strings"

	"github.com/abyssparanoia/goerr"
)

var errFormatMismatch = errors.New("format mismatch")

func someAction(tasks []task) error {
	for _, t := range tasks {
		if err := validateData(t.Data); err != nil {
			return goerr.Wrap(err, "failed to validate data").WithValue("name", t.Name)
		}
	}
	// ....
	return nil
}

func validateData(data string) error {
	if !strings.HasPrefix(data, "data:") {
		return goerr.Wrap(errFormatMismatch).WithValue("data", data)
	}
	return nil
}

type task struct {
	Name string
	Data string
}

func main() {
	tasks := []task{
		{Name: "task1", Data: "data:1"},
		{Name: "task2", Data: "invalid"},
		{Name: "task3", Data: "data:3"},
	}
	if err := someAction(tasks); err != nil {
		if goErr := goerr.Unwrap(err); goErr != nil {
			for k, v := range goErr.Values() {
				log.Printf("var: %s => %v\n", k, v)
			}
		}
		log.Fatalf("msg: %s", err)
	}
}
