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

func TestJoin(t *testing.T) {
	err1 := fmt.Errorf("first error")
	err2 := fmt.Errorf("second error")

	errs := goerr.Join(err1, err2)
	if errs == nil {
		t.Error("Join should return non-nil for valid errors")
	}
	if errs.Len() != 2 {
		t.Errorf("Expected 2 errors, got %d", errs.Len())
	}

	// Test with nil errors
	errs = goerr.Join(nil, nil)
	if errs != nil {
		t.Error("Join should return nil for all nil errors")
	}
}

func TestAppend(t *testing.T) {
	var errs *goerr.Errors
	err1 := fmt.Errorf("first error")
	err2 := fmt.Errorf("second error")

	// Test nil base
	errs = goerr.Append(errs, err1)
	if errs == nil {
		t.Error("Append should create new Errors for nil base")
	}
	if errs.Len() != 1 {
		t.Errorf("Expected 1 error, got %d", errs.Len())
	}

	// Test append to existing
	errs = goerr.Append(errs, err2)
	if errs.Len() != 2 {
		t.Errorf("Expected 2 errors, got %d", errs.Len())
	}
}

func TestAsErrors(t *testing.T) {
	err1 := fmt.Errorf("standard error")
	errs := goerr.Join(err1)

	// Test extracting Errors
	extracted := goerr.AsErrors(errs)
	if extracted == nil {
		t.Error("AsErrors should extract Errors successfully")
	}

	// Test with non-Errors type
	extracted = goerr.AsErrors(err1)
	if extracted != nil {
		t.Error("AsErrors should return nil for non-Errors type")
	}
}

func TestErrorsErrorMethod(t *testing.T) {
	err1 := fmt.Errorf("first error")
	err2 := fmt.Errorf("second error")

	// Test single error
	errs := goerr.Join(err1)
	if errs.Error() != "first error" {
		t.Errorf("Expected 'first error', got '%s'", errs.Error())
	}

	// Test multiple errors
	errs = goerr.Join(err1, err2)
	expected := "first error\nsecond error"
	if errs.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, errs.Error())
	}

	// Test nil Errors
	var nilErrs *goerr.Errors
	if nilErrs.Error() != "" {
		t.Error("Nil Errors should return empty string")
	}
}

func TestErrorsUnwrap(t *testing.T) {
	err1 := fmt.Errorf("first error")
	err2 := fmt.Errorf("second error")

	errs := goerr.Join(err1, err2)
	unwrapped := errs.Unwrap()

	if len(unwrapped) != 2 {
		t.Errorf("Expected 2 unwrapped errors, got %d", len(unwrapped))
	}

	if unwrapped[0].Error() != "first error" {
		t.Errorf("Expected 'first error', got '%s'", unwrapped[0].Error())
	}

	if unwrapped[1].Error() != "second error" {
		t.Errorf("Expected 'second error', got '%s'", unwrapped[1].Error())
	}
}

func TestErrorsIs(t *testing.T) {
	err1 := fmt.Errorf("first error")
	err2 := fmt.Errorf("second error")
	err3 := fmt.Errorf("third error")

	errs := goerr.Join(err1, err2)

	if !errors.Is(errs, err1) {
		t.Error("errors.Is should find err1 in Errors")
	}

	if !errors.Is(errs, err2) {
		t.Error("errors.Is should find err2 in Errors")
	}

	if errors.Is(errs, err3) {
		t.Error("errors.Is should not find err3 in Errors")
	}
}

func TestErrorsAs(t *testing.T) {
	customErr := &CustomError{msg: "custom error"}
	standardErr := fmt.Errorf("standard error")

	errs := goerr.Join(customErr, standardErr)

	var target *CustomError
	if !errors.As(errs, &target) {
		t.Error("errors.As should find CustomError in Errors")
	}

	if target.msg != "custom error" {
		t.Errorf("Expected 'custom error', got '%s'", target.msg)
	}
}

func TestErrorsConvenienceMethods(t *testing.T) {
	var nilErrs *goerr.Errors

	// Test nil safety
	if !nilErrs.IsEmpty() {
		t.Error("Nil Errors should be empty")
	}
	if nilErrs.Len() != 0 {
		t.Error("Nil Errors should have length 0")
	}
	if nilErrs.ErrorOrNil() != nil {
		t.Error("Nil Errors should return nil from ErrorOrNil")
	}

	// Test with actual errors
	err1 := fmt.Errorf("test error")
	errs := goerr.Join(err1)

	if errs.IsEmpty() {
		t.Error("Non-empty Errors should not be empty")
	}
	if errs.Len() != 1 {
		t.Errorf("Expected length 1, got %d", errs.Len())
	}
	if errs.ErrorOrNil() == nil {
		t.Error("Non-empty Errors should return self from ErrorOrNil")
	}
}

func TestErrorsMarshalJSON(t *testing.T) {
	err1 := fmt.Errorf("first error")
	err2 := fmt.Errorf("second error")
	errs := goerr.Join(err1, err2)

	data, err := json.Marshal(errs)
	if err != nil {
		t.Fatalf("Failed to marshal Errors: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	errorsArray, ok := result["errors"].([]any)
	if !ok {
		t.Error("Expected 'errors' field to be array")
	}
	if len(errorsArray) != 2 {
		t.Errorf("Expected 2 errors in JSON, got %d", len(errorsArray))
	}
}

func TestErrorsMarshalJSONWithFailingMarshaler(t *testing.T) {
	// Create an error that implements json.Marshaler but fails
	failingErr := &FailingMarshaler{msg: "failing error"}
	normalErr := fmt.Errorf("normal error")

	errs := goerr.Join(failingErr, normalErr)

	data, err := json.Marshal(errs)
	if err != nil {
		t.Fatalf("Failed to marshal Errors: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	errorsArray, ok := result["errors"].([]any)
	if !ok {
		t.Error("Expected 'errors' field to be array")
	}
	if len(errorsArray) != 2 {
		t.Errorf("Expected 2 errors in JSON, got %d", len(errorsArray))
	}

	// First error should fallback to string representation
	if errorsArray[0] != "failing error" {
		t.Errorf("Expected failing marshaler to fallback to string, got %v", errorsArray[0])
	}
}

func TestErrorsLogValue(t *testing.T) {
	err1 := fmt.Errorf("first error")
	err2 := fmt.Errorf("second error")
	errs := goerr.Join(err1, err2)

	// Test that LogValue returns a group
	logValue := errs.LogValue()
	if logValue.Kind() != slog.KindGroup {
		t.Error("LogValue should return a group")
	}
	
	// Verify the structure of the LogValue directly
	groupAttrs := logValue.Group()
	
	// Verify we have the expected attributes: count and errors
	var countAttr, errorsGroupAttr *slog.Attr
	for i, attr := range groupAttrs {
		switch attr.Key {
		case "count":
			countAttr = &groupAttrs[i]
		case "errors":
			errorsGroupAttr = &groupAttrs[i]
		}
	}
	
	// Verify count attribute
	if countAttr == nil {
		t.Error("Missing 'count' attribute in errors group")
	} else {
		count := countAttr.Value.Any()
		if count != int64(2) {
			t.Errorf("Expected count 2, got %v (type: %T)", count, count)
		}
	}
	
	// Verify errors group attribute
	if errorsGroupAttr == nil {
		t.Error("Missing 'errors' group attribute")
	} else if errorsGroupAttr.Value.Kind() != slog.KindGroup {
		t.Error("Errors group should be a group")
	} else {
		// Verify the individual error entries have proper key-value structure
		errorGroupAttrs := errorsGroupAttr.Value.Group()
		
		// Should have 2 attributes: "0" and "1" (the keys), each with their error message values
		if len(errorGroupAttrs) != 2 {
			t.Errorf("Expected 2 attributes in errors group (key-value pairs), got %d", len(errorGroupAttrs))
			// Debug: print what we got
			for i, attr := range errorGroupAttrs {
				t.Logf("Attr[%d]: Key='%s', Value=%v", i, attr.Key, attr.Value.Any())
			}
		} else {
			// Verify first error (key "0")
			if errorGroupAttrs[0].Key != "0" {
				t.Errorf("Expected first key to be '0', got '%s'", errorGroupAttrs[0].Key)
			}
			if errorGroupAttrs[0].Value.Any() != "first error" {
				t.Errorf("Expected first error value to be 'first error', got %v", errorGroupAttrs[0].Value.Any())
			}
			
			// Verify second error (key "1")
			if errorGroupAttrs[1].Key != "1" {
				t.Errorf("Expected second key to be '1', got '%s'", errorGroupAttrs[1].Key)
			}
			if errorGroupAttrs[1].Value.Any() != "second error" {
				t.Errorf("Expected second error value to be 'second error', got %v", errorGroupAttrs[1].Value.Any())
			}
		}
	}
	
	// Test that the LogValue is properly structured for the slog.Group fix
	// This test verifies that we're using proper key-value pairs in slog.Group
	// which would have been broken before the fix
	
	// Verify that errors group contains key-value pairs as expected
	if errorsGroupAttr != nil && errorsGroupAttr.Value.Kind() == slog.KindGroup {
		errorGroupAttrs := errorsGroupAttr.Value.Group()
		
		// This is the critical test: the slog.Group should have proper key-value pairs
		// Before the fix, this would have been malformed
		for i, attr := range errorGroupAttrs {
			// Each attribute should have a string key (the error index)
			expectedKey := fmt.Sprintf("%d", i)
			if attr.Key != expectedKey {
				t.Errorf("Error group attribute %d should have key '%s', got '%s'", i, expectedKey, attr.Key)
			}
			
			// Each value should be a string (the error message)
			if attr.Value.Kind() != slog.KindString {
				t.Errorf("Error group attribute %d should have string value, got %v", i, attr.Value.Kind())
			}
		}
	}
}

func TestErrorsLogValueWithGoerr(t *testing.T) {
	// Test with goerr.Error instances to ensure LogValue is processed correctly
	tag := goerr.NewTag("test-tag")
	err1 := goerr.New("first goerr error", goerr.Tag(tag), goerr.Value("key1", "value1"))
	err2 := goerr.New("second goerr error", goerr.Value("key2", "value2"))
	errs := goerr.Join(err1, err2)

	logValue := errs.LogValue()
	if logValue.Kind() != slog.KindGroup {
		t.Error("LogValue should return a group")
	}
	
	groupAttrs := logValue.Group()
	
	// Find the errors group
	var errorsGroupAttr *slog.Attr
	for i, attr := range groupAttrs {
		if attr.Key == "errors" {
			errorsGroupAttr = &groupAttrs[i]
			break
		}
	}
	
	if errorsGroupAttr == nil {
		t.Fatal("Missing 'errors' group attribute")
	}
	
	// Verify the errors group contains LogValue from goerr.Error instances
	if errorsGroupAttr.Value.Kind() != slog.KindGroup {
		t.Error("Errors group should be a group")
	} else {
		errorGroupAttrs := errorsGroupAttr.Value.Group()
		
		// Should have 2 entries for the 2 errors
		if len(errorGroupAttrs) != 2 {
			t.Errorf("Expected 2 error entries, got %d", len(errorGroupAttrs))
		}
		
		// First error should be processed through LogValue (returns a group)
		if len(errorGroupAttrs) >= 1 {
			if errorGroupAttrs[0].Key != "0" {
				t.Errorf("First error key should be '0', got '%s'", errorGroupAttrs[0].Key)
			}
			// Since err1 implements LogValuer, it should return a group
			if errorGroupAttrs[0].Value.Kind() != slog.KindGroup {
				t.Error("First error should be processed through LogValue and return a group")
			}
		}
		
		// Second error should also be processed through LogValue
		if len(errorGroupAttrs) >= 2 {
			if errorGroupAttrs[1].Key != "1" {
				t.Errorf("Second error key should be '1', got '%s'", errorGroupAttrs[1].Key)
			}
			// Since err2 implements LogValuer, it should return a group
			if errorGroupAttrs[1].Value.Kind() != slog.KindGroup {
				t.Error("Second error should be processed through LogValue and return a group")
			}
		}
	}
}


func TestErrorsFormat(t *testing.T) {
	err1 := fmt.Errorf("first error")
	err2 := fmt.Errorf("second error")
	errs := goerr.Join(err1, err2)

	// Test %v format
	result := fmt.Sprintf("%v", errs)
	expected := "first error\nsecond error"
	if result != expected {
		t.Errorf("Expected '%s' in format, got '%s'", expected, result)
	}

	// Test %+v format
	detailedResult := fmt.Sprintf("%+v", errs)
	if !strings.Contains(detailedResult, "Errors (2)") {
		t.Errorf("Expected 'Errors (2)' in detailed format, got '%s'", detailedResult)
	}
}

func TestErrorsFlatten(t *testing.T) {
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")
	err3 := fmt.Errorf("error 3")

	// Create nested Errors
	inner := goerr.Join(err1, err2)
	outer := goerr.Append(nil, inner, err3)

	// Should flatten to 3 individual errors
	if outer.Len() != 3 {
		t.Errorf("Expected 3 flattened errors, got %d", outer.Len())
	}

	unwrapped := outer.Unwrap()
	if len(unwrapped) != 3 {
		t.Errorf("Expected 3 unwrapped errors, got %d", len(unwrapped))
	}
}

func TestErrorsHasTag(t *testing.T) {
	tag1 := goerr.NewTag("tag1")
	tag2 := goerr.NewTag("tag2")
	tag3 := goerr.NewTag("tag3")

	// Create errors with tags
	err1 := goerr.New("error 1", goerr.Tag(tag1))
	err2 := goerr.New("error 2", goerr.Tag(tag2))
	err3 := fmt.Errorf("error 3") // no tag

	// Test Errors with tagged errors
	errs := goerr.Join(err1, err2, err3)

	if !errs.HasTag(tag1) {
		t.Error("Errors should have tag1 from err1")
	}

	if !errs.HasTag(tag2) {
		t.Error("Errors should have tag2 from err2")
	}

	if errs.HasTag(tag3) {
		t.Error("Errors should not have tag3")
	}

	// Test with nil Errors
	var nilErrs *goerr.Errors
	if nilErrs.HasTag(tag1) {
		t.Error("Nil Errors should not have any tags")
	}

	// Test with empty Errors
	emptyErrs := goerr.Join()
	if emptyErrs != nil && emptyErrs.HasTag(tag1) {
		t.Error("Empty Errors should not have any tags")
	}

	// Test goerr.HasTag function with Errors
	if !goerr.HasTag(errs, tag1) {
		t.Error("goerr.HasTag should find tag1 in Errors")
	}

	if !goerr.HasTag(errs, tag2) {
		t.Error("goerr.HasTag should find tag2 in Errors")
	}

	if goerr.HasTag(errs, tag3) {
		t.Error("goerr.HasTag should not find tag3 in Errors")
	}
}

func TestErrorsHasTagNested(t *testing.T) {
	tag1 := goerr.NewTag("nested-tag")
	tag2 := goerr.NewTag("outer-tag")

	// Create nested structure
	innerErr := goerr.New("inner error", goerr.Tag(tag1))
	innerErrs := goerr.Join(innerErr)

	outerErr := goerr.New("outer error", goerr.Tag(tag2))
	allErrs := goerr.Join(innerErrs, outerErr)

	// Should find tags from nested Errors
	if !allErrs.HasTag(tag1) {
		t.Error("Should find tag from nested Errors")
	}

	if !allErrs.HasTag(tag2) {
		t.Error("Should find tag from direct error")
	}

	unknownTag := goerr.NewTag("unknown")
	if allErrs.HasTag(unknownTag) {
		t.Error("Should not find unknown tag")
	}
}

// CustomError for testing errors.As
type CustomError struct {
	msg string
}

func (e *CustomError) Error() string {
	return e.msg
}

// FailingMarshaler for testing JSON marshaling error handling
type FailingMarshaler struct {
	msg string
}

func (e *FailingMarshaler) Error() string {
	return e.msg
}

func (e *FailingMarshaler) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("marshaling failed")
}
