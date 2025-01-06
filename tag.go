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
//			return nil, goerr.New("user not found").WithTags(TagNotFound)
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
type Tag struct {
	value string
}

// NewTag creates a new Tag. The key will be empty.
func NewTag(value string) Tag {
	return Tag{value: value}
}

// String returns the string representation of the Tag. It's for implementing fmt.Stringer interface.
func (t Tag) String() string {
	return t.value
}

// Format writes the Tag to the writer. It's for implementing fmt.Formatter interface.
func (t Tag) Format(s fmt.State, verb rune) {
	_, _ = io.WriteString(s, t.value)
}

// WithTags adds tags to the error. The tags are used to categorize errors.
func (x *Error) WithTags(tags ...Tag) *Error {
	for _, tag := range tags {
		x.tags.add(tag)
	}
	return x
}

// HasTag returns true if the error has the tag.
func (x *Error) HasTag(tag Tag) bool {
	return x.tags.has(tag)
}

type tags map[Tag]struct{}

func (t tags) add(tag Tag) {
	t[tag] = struct{}{}
}

func (t tags) has(tag Tag) bool {
	_, ok := t[tag]
	return ok
}

func (t tags) clone() tags {
	newTags := make(tags)
	for tag := range t {
		newTags[tag] = struct{}{}
	}
	return newTags
}
