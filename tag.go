package goerr

import (
	"fmt"
	"io"
)

// TagKey is a type to represent a key of the tag. The struct should be created by only NewTagKey function.
type TagKey struct {
	key string
}

func NewTagKey(key string) TagKey {
	return TagKey{key: key}
}

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
	key   TagKey
	value string
}

// NewTag creates a new Tag. The key will be empty.
func NewTag(value string) Tag {
	return Tag{value: value}
}

// NewTagWithKey creates a new Tag with the key. The key is used to identify the tag.
func NewTagWithKey(key TagKey, value string) Tag {
	return Tag{key: key, value: value}
}

// String returns the string representation of the Tag. It's for implementing fmt.Stringer interface.
func (t Tag) String() string {
	return t.value
}

// Format writes the Tag to the writer. It's for implementing fmt.Formatter interface.
func (t Tag) Format(s fmt.State, verb rune) {
	io.WriteString(s, t.value)
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

// LookupTag returns the tag with the key. If the tag is not found, it returns false.
func (x *Error) LookupTag(key TagKey) (Tag, bool) {
	tag, ok := x.tags.tagsWithKey[key]
	return tag, ok
}

type tagStorage struct {
	tagsWithoutKey map[Tag]struct{}
	tagsWithKey    map[TagKey]Tag
}

func newTagStorage() tagStorage {
	return tagStorage{
		tagsWithoutKey: make(map[Tag]struct{}),
		tagsWithKey:    make(map[TagKey]Tag),
	}
}

func (t tagStorage) add(tag Tag) {
	if tag.key.key == "" {
		t.tagsWithoutKey[tag] = struct{}{}
		return
	}
	t.tagsWithKey[tag.key] = tag
}

func (t tagStorage) has(tag Tag) bool {
	if tag.key.key == "" {
		_, ok := t.tagsWithoutKey[tag]
		return ok
	}
	_, ok := t.tagsWithKey[tag.key]
	return ok
}
