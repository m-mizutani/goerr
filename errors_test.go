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

func TestErrorsLogValue(t *testing.T) {
	err1 := fmt.Errorf("first error")
	err2 := fmt.Errorf("second error")
	errs := goerr.Join(err1, err2)

	logValue := errs.LogValue()
	if logValue.Kind() != slog.KindGroup {
		t.Error("LogValue should return a group")
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

// CustomError for testing errors.As
type CustomError struct {
	msg string
}

func (e *CustomError) Error() string {
	return e.msg
}
