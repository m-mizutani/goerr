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
