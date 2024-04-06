package main

import (
	"io/fs"
	"log"
	"os"

	"github.com/m-mizutani/goerr"
)

func firstFunc(label string) error {
	_, err := secondFunc(label+".txt", os.O_RDONLY, 0644)
	if err != nil {
		return goerr.Wrap(err, "failed to call secondFunc").With("label", label)
	}
	// .....
	return nil
}

func secondFunc(fname string, flag int, perm fs.FileMode) ([]byte, error) {
	if _, err := os.OpenFile(fname, flag, perm); err != nil {
		return nil, goerr.Wrap(err).With("fname", fname).With("flag", flag)
	}
	// .....
	return nil, nil
}

func main() {
	if err := firstFunc("no_such_file"); err != nil {
		if goErr := goerr.Unwrap(err); goErr != nil {
			for k, v := range goErr.Values() {
				log.Printf("var: %s => %v\n", k, v)
			}
		}
		log.Fatalf("msg: %s", err)
	}
}
