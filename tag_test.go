package goerr_test

import (
	"fmt"

	"github.com/m-mizutani/goerr"
)

func ExampleNewTag() {
	t1 := goerr.NewTag("DB error")
	err := goerr.New("error message").WithTags(t1)

	if goErr := goerr.Unwrap(err); goErr != nil {
		if goErr.HasTag(t1) {
			fmt.Println("DB error")
		}
	}
	// Output: DB error
}

func ExampleNewTagWithKey() {
	k1 := goerr.NewTagKey("http")
	tagNotFound := goerr.NewTagWithKey(k1, "404")
	tagUnauthorized := goerr.NewTagWithKey(k1, "401")
	err := goerr.New("resource not found").WithTags(tagNotFound)

	if goErr := goerr.Unwrap(err); goErr != nil {
		switch tag, ok := goErr.LookupTag(k1); {
		case ok && tag == tagNotFound:
			fmt.Println("not found")
		case ok && tag == tagUnauthorized:
			fmt.Println("unauthorized")
		default:
			fmt.Println("unknown")
		}
	}
	// Output: not found
}
