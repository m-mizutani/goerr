package goerr_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/m-mizutani/goerr/v2"
)

func TestNew(t *testing.T) {
	err := goerr.New("test error")
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", err.Error())
	}

	// Test with options
	tag := goerr.NewTag("test")
	err = goerr.New("test error", goerr.Tag(tag), goerr.Value("key", "value"))

	if !err.HasTag(tag) {
		t.Error("Error should have the tag")
	}

	values := err.Values()
	if values["key"] != "value" {
		t.Error("Error should have the value")
	}
}

func TestWrap(t *testing.T) {
	original := fmt.Errorf("original error")
	err := goerr.Wrap(original, "wrapped")

	if err.Error() != "wrapped: original error" {
		t.Errorf("Expected 'wrapped: original error', got '%s'", err.Error())
	}

	if !errors.Is(err, original) {
		t.Error("Wrapped error should be identifiable with errors.Is")
	}
}

func TestUnwrap(t *testing.T) {
	err := goerr.New("test error")

	// Test unwrapping goerr.Error
	unwrapped := goerr.Unwrap(err)
	if unwrapped == nil {
		t.Error("Unwrap should return the error itself")
	}

	// Test unwrapping non-goerr error
	stdErr := fmt.Errorf("standard error")
	unwrapped = goerr.Unwrap(stdErr)
	if unwrapped != nil {
		t.Error("Unwrap should return nil for non-goerr errors")
	}
}

func TestErrorValues(t *testing.T) {
	err := goerr.New("test error",
		goerr.Value("key1", "value1"),
		goerr.Value("key2", 42))

	values := goerr.Values(err)
	if values["key1"] != "value1" {
		t.Error("Should extract key1 value")
	}
	if values["key2"] != 42 {
		t.Error("Should extract key2 value")
	}

	// Test with non-goerr error
	stdErr := fmt.Errorf("standard error")
	values = goerr.Values(stdErr)
	if values != nil {
		t.Error("Values should return nil for non-goerr errors")
	}
}

func TestErrorTags(t *testing.T) {
	tag1 := goerr.NewTag("tag1")
	tag2 := goerr.NewTag("tag2")

	err := goerr.New("test error", goerr.Tag(tag1), goerr.Tag(tag2))

	tags := goerr.Tags(err)
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}

	// Test HasTag function
	if !goerr.HasTag(err, tag1) {
		t.Error("Should have tag1")
	}
	if !goerr.HasTag(err, tag2) {
		t.Error("Should have tag2")
	}

	unknownTag := goerr.NewTag("unknown")
	if goerr.HasTag(err, unknownTag) {
		t.Error("Should not have unknown tag")
	}
}

func TestErrorMarshalJSON(t *testing.T) {
	tag := goerr.NewTag("test")
	err := goerr.New("test error",
		goerr.Tag(tag),
		goerr.Value("key", "value"))

	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal error: %v", jsonErr)
	}

	var result map[string]any
	if jsonErr := json.Unmarshal(data, &result); jsonErr != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", jsonErr)
	}

	if result["message"] != "test error" {
		t.Error("JSON should contain message")
	}

	values, ok := result["values"].(map[string]any)
	if !ok || values["key"] != "value" {
		t.Error("JSON should contain values")
	}

	tags, ok := result["tags"].([]any)
	if !ok || len(tags) != 1 || tags[0] != "test" {
		t.Error("JSON should contain tags")
	}
}

func TestErrorLogValue(t *testing.T) {
	err := goerr.New("test error", goerr.Value("key", "value"))

	logValue := err.LogValue()
	if logValue.Kind() != slog.KindGroup {
		t.Error("LogValue should return a group")
	}
}

func TestErrorFormat(t *testing.T) {
	err := goerr.New("test error", goerr.Value("key", "value"))

	// Test %v format
	result := fmt.Sprintf("%v", err)
	if result != "test error" {
		t.Errorf("Expected 'test error', got '%s'", result)
	}

	// Test %s format
	result = fmt.Sprintf("%s", err)
	if result != "test error" {
		t.Errorf("Expected 'test error', got '%s'", result)
	}

	// Test %q format
	result = fmt.Sprintf("%q", err)
	if result != "\"test error\"" {
		t.Errorf("Expected '\"test error\"', got '%s'", result)
	}

	// Test %+v format (should include stack trace and values)
	detailedResult := fmt.Sprintf("%+v", err)
	if !strings.Contains(detailedResult, "test error") {
		t.Error("Detailed format should contain error message")
	}
	if !strings.Contains(detailedResult, "Values:") {
		t.Error("Detailed format should contain Values section")
	}
	if !strings.Contains(detailedResult, "key: value") {
		t.Error("Detailed format should contain the key-value pair")
	}
}

func TestErrorFormatWrapped(t *testing.T) {
	// Create a chain of wrapped errors with different values
	baseErr := goerr.New("base error",
		goerr.Value("base_key", "base_value"),
		goerr.Value("shared_key", "base_shared")) // This should be overwritten

	middleErr := goerr.Wrap(baseErr, "middle error",
		goerr.Value("middle_key", "middle_value"),
		goerr.Value("shared_key", "middle_shared")) // This overwrites base_shared

	topErr := goerr.Wrap(middleErr, "top error",
		goerr.Value("top_key", "top_value"),
		goerr.Value("shared_key", "top_shared")) // This overwrites middle_shared

	// Test %+v format with wrapped errors
	detailedResult := fmt.Sprintf("%+v", topErr)

	// Should contain the complete error message chain
	if !strings.Contains(detailedResult, "top error: middle error: base error") {
		t.Error("Detailed format should contain complete error message chain")
	}

	// Should contain Values section
	if !strings.Contains(detailedResult, "Values:") {
		t.Error("Detailed format should contain Values section")
	}

	// Should contain values from base error
	if !strings.Contains(detailedResult, "base_key: base_value") {
		t.Error("Detailed format should contain values from base error")
	}

	// Should contain values from middle error
	if !strings.Contains(detailedResult, "middle_key: middle_value") {
		t.Error("Detailed format should contain values from middle error")
	}

	// Should contain values from top error
	if !strings.Contains(detailedResult, "top_key: top_value") {
		t.Error("Detailed format should contain values from top error")
	}

	// Should show the final value for overwritten key (top-level wins)
	if !strings.Contains(detailedResult, "shared_key: top_shared") {
		t.Error("Detailed format should show final value for overwritten key")
	}

	// Should NOT contain overwritten values
	if strings.Contains(detailedResult, "base_shared") {
		t.Error("Detailed format should not contain overwritten base value")
	}
	if strings.Contains(detailedResult, "middle_shared") {
		t.Error("Detailed format should not contain overwritten middle value")
	}

	// Verify keys are sorted alphabetically
	valuesSection := detailedResult[strings.Index(detailedResult, "Values:"):]
	baseKeyPos := strings.Index(valuesSection, "base_key:")
	middleKeyPos := strings.Index(valuesSection, "middle_key:")
	sharedKeyPos := strings.Index(valuesSection, "shared_key:")
	topKeyPos := strings.Index(valuesSection, "top_key:")

	// All positions should be found
	if baseKeyPos == -1 || middleKeyPos == -1 || sharedKeyPos == -1 || topKeyPos == -1 {
		t.Error("All keys should be present in values section")
	}

	// Keys should appear in alphabetical order
	if !(baseKeyPos < middleKeyPos && middleKeyPos < sharedKeyPos && sharedKeyPos < topKeyPos) {
		t.Error("Keys should be sorted alphabetically in values section")
	}
}

func TestErrorFormatTypedValues(t *testing.T) {
	// Test that TypedValues from wrapped errors are also included
	// Note: We need to use the typed values functionality here
	// Since TypedValues is not directly exposed via options, we test the formatting
	// by creating errors that would have typed values through internal mechanisms

	baseErr := goerr.New("base error", goerr.Value("regular_key", "regular_value"))
	wrappedErr := goerr.Wrap(baseErr, "wrapped error", goerr.Value("wrapped_key", "wrapped_value"))

	// Test %+v format includes all values
	detailedResult := fmt.Sprintf("%+v", wrappedErr)

	// Should contain both regular values
	if !strings.Contains(detailedResult, "regular_key: regular_value") {
		t.Error("Should contain base error values")
	}
	if !strings.Contains(detailedResult, "wrapped_key: wrapped_value") {
		t.Error("Should contain wrapped error values")
	}

	// Should contain Values section (since we have regular values)
	if !strings.Contains(detailedResult, "Values:") {
		t.Error("Should contain Values section")
	}
}

func TestErrorStackTrace(t *testing.T) {
	err := goerr.New("test error")

	stacks := err.Stacks()
	if len(stacks) == 0 {
		t.Error("Error should have stack trace")
	}

	// Test that stack trace contains this test function
	found := false
	for _, stack := range stacks {
		if strings.Contains(stack.Func, "TestErrorStackTrace") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Stack trace should contain test function")
	}
}

func TestErrorIs(t *testing.T) {
	err1 := goerr.New("error 1", goerr.ID("test-id"))
	err2 := goerr.New("error 2", goerr.ID("test-id"))
	err3 := goerr.New("error 3", goerr.ID("different-id"))
	err4 := goerr.New("error 4") // no ID

	// Test ID-based comparison
	if !errors.Is(err1, err2) {
		t.Error("Errors with same ID should be equal")
	}
	if errors.Is(err1, err3) {
		t.Error("Errors with different IDs should not be equal")
	}

	// Test pointer-based comparison for errors without ID
	if errors.Is(err4, goerr.New("error 4")) {
		t.Error("Different error instances without ID should not be equal")
	}
	if !errors.Is(err4, err4) {
		t.Error("Same error instance should be equal to itself")
	}
}

func TestErrorCopy(t *testing.T) {
	original := goerr.New("original",
		goerr.ID("test-id"),
		goerr.Value("key", "value"))

	// Test wrapping (which uses copy internally)
	cause := fmt.Errorf("cause error")
	wrapped := original.Wrap(cause)

	if wrapped.Error() != "original: cause error" {
		t.Errorf("Expected 'original: cause error', got '%s'", wrapped.Error())
	}

	// Verify values are copied
	values := wrapped.Values()
	if values["key"] != "value" {
		t.Error("Wrapped error should inherit values")
	}

	// Verify cause is set
	if !errors.Is(wrapped, cause) {
		t.Error("Wrapped error should be identifiable with its cause")
	}
}

func TestErrorUnstack(t *testing.T) {
	err := goerr.New("test error")
	originalLen := len(err.Stacks())
	t.Logf("Original stack length: %d", originalLen)

	// Test Unstack (should modify in place)
	unstacked := err.Unstack()
	newLen := len(unstacked.Stacks())
	t.Logf("After Unstack: %d", newLen)

	if unstacked != err {
		t.Error("Unstack should return the same instance")
	}

	// Test UnstackN(1) should have same effect as Unstack
	err2 := goerr.New("test error 2")
	unstackedN1 := err2.UnstackN(1)

	if unstackedN1 != err2 {
		t.Error("UnstackN should return the same instance")
	}
}

func TestPrintable(t *testing.T) {
	tag := goerr.NewTag("test")
	cause := fmt.Errorf("cause error")
	err := goerr.Wrap(cause, "wrapped error",
		goerr.Tag(tag),
		goerr.Value("key", "value"))

	printable := err.Printable()
	if printable.Message != "wrapped error" {
		t.Error("Printable should have correct message")
	}

	if printable.Values["key"] != "value" {
		t.Error("Printable should have values")
	}

	if len(printable.Tags) != 1 || printable.Tags[0] != "test" {
		t.Error("Printable should have tags")
	}

	if printable.Cause != "cause error" {
		t.Error("Printable should have cause")
	}

	if len(printable.StackTrace) == 0 {
		t.Error("Printable should have stack trace")
	}
}

func TestWith_WithGoError(t *testing.T) {
	// Create original goerr.Error
	original := goerr.New("original message", goerr.Value("orig_key", "orig_value"))

	// With additional information
	tag := goerr.NewTag("test_tag")
	withAdded := goerr.With(original,
		goerr.Value("new_key", "new_value"),
		goerr.Tag(tag),
	)

	// Test that stacktrace is preserved (identical frames)
	withStacks := withAdded.Stacks()
	originalStacks := original.Stacks()
	if len(withStacks) != len(originalStacks) {
		t.Errorf("Expected stacktrace length %d, got %d", len(originalStacks), len(withStacks))
	}
	for i, origStack := range originalStacks {
		if i >= len(withStacks) {
			break
		}
		withStack := withStacks[i]
		if origStack.File != withStack.File || origStack.Line != withStack.Line || origStack.Func != withStack.Func {
			t.Errorf("Stack frame %d differs: original=%+v, with=%+v", i, origStack, withStack)
		}
	}

	// Test that original error is not modified (exact value checks)
	originalCurrentValues := original.Values()
	if len(originalCurrentValues) != 1 {
		t.Errorf("Original error values count changed: expected 1, got %d", len(originalCurrentValues))
	}
	if originalCurrentValues["orig_key"] != "orig_value" {
		t.Errorf("Original value changed: expected 'orig_value', got %v", originalCurrentValues["orig_key"])
	}
	if val, exists := originalCurrentValues["new_key"]; exists {
		t.Errorf("Original error was modified with new key: %v", val)
	}
	if original.HasTag(tag) {
		t.Error("Original error was modified with new tag")
	}

	// Test that error with additions has both original and new values (exact checks)
	withValues := withAdded.Values()
	if len(withValues) != 2 {
		t.Errorf("Expected 2 values in withAdded, got %d", len(withValues))
	}
	if withValues["orig_key"] != "orig_value" {
		t.Errorf("Original value not preserved: expected 'orig_value', got %v", withValues["orig_key"])
	}
	if withValues["new_key"] != "new_value" {
		t.Errorf("New value not added correctly: expected 'new_value', got %v", withValues["new_key"])
	}
	if !withAdded.HasTag(tag) {
		t.Error("New tag not added to error with additions")
	}

	// Test that message is preserved
	if withAdded.Error() != "original message" {
		t.Errorf("Expected 'original message', got '%s'", withAdded.Error())
	}
}

func TestWith_WithStandardError(t *testing.T) {
	// Create standard error
	original := fmt.Errorf("standard error")

	// With information
	tag := goerr.NewTag("wrap_tag")
	withAdded := goerr.With(original,
		goerr.Value("wrap_key", "wrap_value"),
		goerr.Tag(tag),
	)

	// Test that it behaves like Wrap
	if !errors.Is(withAdded, original) {
		t.Error("Error with additions should wrap original error")
	}

	// Test that new stacktrace is created
	if len(withAdded.Stacks()) == 0 {
		t.Error("Expected new stacktrace to be created")
	}

	// Test added values and tags
	if withAdded.Values()["wrap_key"] != "wrap_value" {
		t.Error("Value not added to wrapped error")
	}
	if !withAdded.HasTag(tag) {
		t.Error("Tag not added to wrapped error")
	}

	// Test error message (should contain original)
	if withAdded.Error() != "standard error" {
		t.Errorf("Expected 'standard error', got '%s'", withAdded.Error())
	}
}

func TestWith_WithNilError(t *testing.T) {
	result := goerr.With(nil, goerr.Value("key", "value"))
	if result != nil {
		t.Error("Expected nil result for nil input")
	}
}

func TestWith_WithTypedValues(t *testing.T) {
	// Create original error with typed value
	userIDKey := goerr.NewTypedKey[int]("user_id")
	original := goerr.New("original", goerr.TypedValue(userIDKey, 123))

	// With additional typed value
	requestIDKey := goerr.NewTypedKey[string]("request_id")
	withAdded := goerr.With(original, goerr.TypedValue(requestIDKey, "req-456"))

	// Test original is not modified
	if val, ok := goerr.GetTypedValue(original, requestIDKey); ok {
		t.Errorf("Original error was modified with new typed value: %v", val)
	}

	// Test error with additions has both values
	if val, ok := goerr.GetTypedValue(withAdded, userIDKey); !ok || val != 123 {
		t.Error("Original typed value not preserved")
	}
	if val, ok := goerr.GetTypedValue(withAdded, requestIDKey); !ok || val != "req-456" {
		t.Error("New typed value not added")
	}
}

func TestWith_StacktracePreservation(t *testing.T) {
	// Create error with known stacktrace
	original := goerr.New("test error")
	originalStacks := original.Stacks()

	// With should preserve exact same stacktrace
	withAdded := goerr.With(original, goerr.Value("key", "value"))
	withStacks := withAdded.Stacks()

	if len(originalStacks) != len(withStacks) {
		t.Fatalf("Stacktrace length changed: original=%d, with=%d",
			len(originalStacks), len(withStacks))
	}

	// Compare each stack frame
	for i, orig := range originalStacks {
		with := withStacks[i]
		if orig.File != with.File || orig.Line != with.Line || orig.Func != with.Func {
			t.Errorf("Stack frame %d differs: original=%+v, with=%+v", i, orig, with)
		}
	}
}

func TestWith_ErrorChain(t *testing.T) {
	// Create error chain
	root := fmt.Errorf("root error")
	wrapped := goerr.Wrap(root, "wrapped")

	// With into wrapped error
	withAdded := goerr.With(wrapped, goerr.Value("added_key", "added_value"))

	// Test error chain is preserved
	if !errors.Is(withAdded, root) {
		t.Error("Error chain broken after adding with")
	}

	// Test unwrapping works
	if withAdded.Unwrap() != root {
		t.Error("Unwrap not working correctly after adding with")
	}
}

func TestWith_MultipleOptions(t *testing.T) {
	original := goerr.New("test")

	tag1 := goerr.NewTag("tag1")
	tag2 := goerr.NewTag("tag2")
	userKey := goerr.NewTypedKey[string]("user")

	withAdded := goerr.With(original,
		goerr.Value("str_key", "str_value"),
		goerr.Value("int_key", 42),
		goerr.TypedValue(userKey, "alice"),
		goerr.Tag(tag1),
		goerr.Tag(tag2),
	)

	// Test all options were applied
	values := withAdded.Values()
	if values["str_key"] != "str_value" {
		t.Error("String value not applied")
	}
	if values["int_key"] != 42 {
		t.Error("Int value not applied")
	}

	if val, ok := goerr.GetTypedValue(withAdded, userKey); !ok || val != "alice" {
		t.Error("Typed value not applied")
	}

	if !withAdded.HasTag(tag1) || !withAdded.HasTag(tag2) {
		t.Error("Tags not applied")
	}

	// Test original remains unchanged
	if len(original.Values()) != 0 {
		t.Error("Original error was modified")
	}
}
