package goerr

import (
	"fmt"
	"io"
)

// Tag is a type to represent an error tag. It is used to categorize errors. The struct should be created by only NewTag function.
//
// Example:
//
//	TagNotFound := NewTag("not_found")
//
//	func FindUser(id string) (*User, error) {
//		...
//		if user == nil {
//			return nil, goerr.New("user not found", goerr.Tag(TagNotFound))
//		}
//		...
//	}
//
//	func main() {
//		err := FindUser("123")
//		if goErr := goerr.Unwrap(err); goErr != nil {
//			if goErr.HasTag(TagNotFound) {
//				fmt.Println("User not found")
//			}
//		}
//	}
type tag struct {
	value string
}

// NewTag creates a new Tag. The key will be empty.
func NewTag(value string) tag {
	return tag{value: value}
}

// String returns the string representation of the Tag. It's for implementing fmt.Stringer interface.
func (t tag) String() string {
	return t.value
}

// Format writes the Tag to the writer. It's for implementing fmt.Formatter interface.
func (t tag) Format(s fmt.State, verb rune) {
	_, _ = io.WriteString(s, t.value)
}

// WithTags adds tags to the error. The tags are used to categorize errors.
//
// Deprecated: Use goerr.Tag instead.
func (x *Error) WithTags(tags ...tag) *Error {
	for _, tag := range tags {
		x.tags[tag] = struct{}{}
	}
	return x
}

// HasTag returns true if the error has the tag.
func (x *Error) HasTag(tag tag) bool {
	tags := x.mergedTags()
	_, ok := tags[tag]
	return ok
}

type tags map[tag]struct{}

func (t tags) clone() tags {
	newTags := make(tags)
	for tag := range t {
		newTags[tag] = struct{}{}
	}
	return newTags
}
