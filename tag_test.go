package goerr_test

import (
	"fmt"
	"testing"

	"github.com/m-mizutani/goerr"
)

func ExampleNewTag() {
	t1 := goerr.NewTag("DB error")
	err := goerr.New("error message", goerr.Tag(t1))

	if goErr := goerr.Unwrap(err); goErr != nil {
		if goErr.HasTag(t1) {
			fmt.Println("DB error")
		}
	}
	// Output: DB error
}

func TestNewTag(t *testing.T) {
	tagValue := "test_tag"
	tag := goerr.NewTag(tagValue)

	if tag.String() != tagValue {
		t.Errorf("expected tag value to be %s, got %s", tagValue, tag.String())
	}
}

func TestWithTags(t *testing.T) {
	tag1 := goerr.NewTag("tag1")
	tag2 := goerr.NewTag("tag2")
	tag3 := goerr.NewTag("tag3")
	err := goerr.New("error message", goerr.Tag(tag1), goerr.Tag(tag2))

	if goErr := goerr.Unwrap(err); goErr != nil {
		if !goErr.HasTag(tag1) {
			t.Errorf("expected error to have tag1")
		}
		if !goErr.HasTag(tag2) {
			t.Errorf("expected error to have tag2")
		}
		if goErr.HasTag(tag3) {
			t.Errorf("expected error to not have tag3")
		}
	}
}

func TestHasTag(t *testing.T) {
	tag := goerr.NewTag("test_tag")
	err := goerr.New("error message", goerr.Tag(tag))

	if goErr := goerr.Unwrap(err); goErr != nil {
		if !goErr.HasTag(tag) {
			t.Errorf("expected error to have tag")
		}
	}
}
