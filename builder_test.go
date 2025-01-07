package goerr_test

import (
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
