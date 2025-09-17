package goerr_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/m-mizutani/goerr/v2"
)

func newErrorWithBuilder() *goerr.Error {
	return goerr.NewBuilder(goerr.V("color", "orange")).New("error")
}

func TestBuilderNew(t *testing.T) {
	err := newErrorWithBuilder()

	if err.Values()["color"] != "orange" {
		t.Errorf("Unexpected value: %v", err.Values())
	}
}

func TestBuilderWrap(t *testing.T) {
	cause := goerr.New("cause")
	err := goerr.NewBuilder(goerr.V("color", "blue")).Wrap(cause, "error")

	if err.Values()["color"] != "blue" {
		t.Errorf("Unexpected value: %v", err.Values())
	}

	if err.Unwrap().Error() != "cause" {
		t.Errorf("Unexpected cause: %v", err.Unwrap().Error())
	}
}

func ExampleNewBuilder() {
	// Create a builder with common context for a request.
	builder := goerr.NewBuilder(
		goerr.Value("service", "auth-service"),
		goerr.Value("request_id", "req-9876"),
	)

	// Use the builder to create errors.
	err1 := builder.New("user not found")
	err2 := builder.Wrap(io.EOF, "failed to read body")

	// The context from the builder is automatically included.
	fmt.Println(goerr.Values(err1)["service"])
	fmt.Println(goerr.Values(err2)["request_id"])

	// Output:
	// auth-service
	// req-9876
}
