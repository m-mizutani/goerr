package goerr_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/m-mizutani/goerr/v2"
)

func oops() *goerr.Error {
	return goerr.New("omg")
}

func normalError() error {
	return fmt.Errorf("red")
}

func wrapError() *goerr.Error {
	err := normalError()
	return goerr.Wrap(err, "orange")
}

func TestNew(t *testing.T) {
	err := oops()
	v := fmt.Sprintf("%+v", err)
	if !strings.Contains(v, "goerr/v2_test.oops") {
		t.Error("Stack trace 'goerr/v2_test.oops' is not found")
	}
	if !strings.Contains(err.Error(), "omg") {
		t.Error("Error message is not correct")
	}
}

func TestOptions(t *testing.T) {
	var testCases = map[string]struct {
		options []goerr.Option
		values  map[string]interface{}
		tags    []string
	}{
		"empty": {
			options: []goerr.Option{},
			values:  map[string]interface{}{},
			tags:    []string{},
		},
		"single value": {
			options: []goerr.Option{goerr.Value("key", "value")},
			values:  map[string]interface{}{"key": "value"},
			tags:    []string{},
		},
		"multiple values": {
			options: []goerr.Option{goerr.Value("key1", "value1"), goerr.Value("key2", "value2")},
			values:  map[string]interface{}{"key1": "value1", "key2": "value2"},
			tags:    []string{},
		},
		"single tag": {
			options: []goerr.Option{goerr.Tag(goerr.NewTag("tag1"))},
			values:  map[string]interface{}{},
			tags:    []string{"tag1"},
		},
		"multiple tags": {
			options: []goerr.Option{goerr.Tag(goerr.NewTag("tag1")), goerr.Tag(goerr.NewTag("tag2"))},
			values:  map[string]interface{}{},
			tags:    []string{"tag1", "tag2"},
		},
		"values and tags": {
			options: []goerr.Option{goerr.Value("key", "value"), goerr.Tag(goerr.NewTag("tag1"))},
			values:  map[string]interface{}{"key": "value"},
			tags:    []string{"tag1"},
		},
	}

	for name, tc := range testCases {
		tc := tc // capture range variable for Go versions < 1.22
		t.Run(name, func(t *testing.T) {
			err := goerr.New("test", tc.options...)
			values := err.Values()
			if len(values) != len(tc.values) {
				t.Errorf("Expected values length to be %d, got %d", len(tc.values), len(values))
			}
			for k, v := range tc.values {
				if values[k] != v {
					t.Errorf("Expected value for key '%s' to be '%v', got '%v'", k, v, values[k])
				}
			}

			tags := goerr.Tags(err)
			if len(tags) != len(tc.tags) {
				t.Errorf("Expected tags length to be %d, got %d", len(tc.tags), len(tags))
			}
			for _, tag := range tc.tags {
				if !sliceHas(tags, tag) {
					t.Errorf("Expected tags to contain '%s'", tag)
				}
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	err := wrapError()
	st := fmt.Sprintf("%+v", err)
	if !strings.Contains(st, "github.com/m-mizutani/goerr/v2_test.wrapError") {
		t.Error("Stack trace 'wrapError' is not found")
	}
	if !strings.Contains(st, "github.com/m-mizutani/goerr/v2_test.TestWrapError") {
		t.Error("Stack trace 'TestWrapError' is not found")
	}
	if strings.Contains(st, "github.com/m-mizutani/goerr/v2_test.normalError") {
		t.Error("Stack trace 'normalError' is found")
	}
	if !strings.Contains(err.Error(), "orange: red") {
		t.Error("Error message is not correct")
	}
}

func TestStackTrace(t *testing.T) {
	err := oops()
	st := err.Stacks()
	if len(st) != 4 {
		t.Errorf("Expected stack length of 4, got %d", len(st))
	}
	if st[0].Func != "github.com/m-mizutani/goerr/v2_test.oops" {
		t.Error("Stack trace 'github.com/m-mizutani/goerr/v2_test.oops' is not found")
	}
	if !regexp.MustCompile(`/goerr/errors_test\.go$`).MatchString(st[0].File) {
		t.Error("Stack trace file is not correct")
	}
	if st[0].Line != 19 {
		t.Errorf("Expected line number 19, got %d", st[0].Line)
	}
}

func TestMultiWrap(t *testing.T) {
	err1 := oops()
	err2 := goerr.Wrap(err1, "some message")
	if err1 == err2 {
		t.Error("Expected err1 and err2 to be different")
	}

	err3 := goerr.Wrap(err1, "some message")
	if err1 == err3 {
		t.Error("Expected err1 and err3 to be different")
	}
}

func TestErrorCode(t *testing.T) {
	rootErr := goerr.New("something bad")
	baseErr1 := goerr.New("oops").ID("code1")
	baseErr2 := goerr.New("oops").ID("code2")

	newErr := baseErr1.Wrap(rootErr, goerr.V("v", 1))

	if !errors.Is(newErr, baseErr1) {
		t.Error("Expected newErr to be based on baseErr1")
	}
	if newErr == baseErr1 {
		t.Error("Expected newErr and baseErr1 to be different")
	}
	if newErr.Values()["v"] == nil {
		t.Error("Expected newErr to have a non-nil value for 'v'")
	}
	if baseErr1.Values()["v"] != nil {
		t.Error("Expected baseErr1 to have a nil value for 'v'")
	}
	if errors.Is(newErr, baseErr2) {
		t.Error("Expected newErr to not be based on baseErr2")
	}
}

func TestPrintableWithGoErr(t *testing.T) {
	cause := errors.New("test")
	err := goerr.Wrap(cause, "oops", goerr.V("blue", "five")).ID("E001")

	p := err.Printable()
	if p.Message != "oops" {
		t.Errorf("Expected message to be 'oops', got '%s'", p.Message)
	}
	if p.ID != "E001" {
		t.Errorf("Expected ID to be 'E001', got '%s'", p.ID)
	}
	if s, ok := p.Cause.(string); !ok {
		t.Errorf("Expected cause is string, got '%t'", p.Cause)
	} else if s != "test" {
		t.Errorf("Expected message is 'test', got '%s'", s)
	}
	if p.Values["blue"] != "five" {
		t.Errorf("Expected value for 'blue' to be 'five', got '%v'", p.Values["blue"])
	}
}

func TestPrintableWithError(t *testing.T) {
	cause := goerr.New("test")
	err := goerr.Wrap(cause, "oops", goerr.V("blue", "five")).ID("E001")

	p := err.Printable()
	if p.Message != "oops" {
		t.Errorf("Expected message to be 'oops', got '%s'", p.Message)
	}
	if p.ID != "E001" {
		t.Errorf("Expected ID to be 'E001', got '%s'", p.ID)
	}
	if cp, ok := p.Cause.(*goerr.Printable); !ok {
		t.Errorf("Expected cause is goerr.Printable, got '%t'", p.Cause)
	} else if cp.Message != "test" {
		t.Errorf("Expected message is 'test', got '%s'", cp.Message)
	}
	if p.Values["blue"] != "five" {
		t.Errorf("Expected value for 'blue' to be 'five', got '%v'", p.Values["blue"])
	}
}

func TestUnwrap(t *testing.T) {
	err1 := goerr.New("omg", goerr.V("color", "five"))
	err2 := fmt.Errorf("oops: %w", err1)

	err := goerr.Unwrap(err2)
	if err == nil {
		t.Error("Expected unwrapped error to be non-nil")
	}
	values := err.Values()
	if values["color"] != "five" {
		t.Errorf("Expected value for 'color' to be 'five', got '%v'", values["color"])
	}
}

func TestErrorString(t *testing.T) {
	err := goerr.Wrap(goerr.Wrap(goerr.New("blue"), "orange"), "red")
	if err.Error() != "red: orange: blue" {
		t.Errorf("Expected error message to be 'red: orange: blue', got '%s'", err.Error())
	}
}

func TestLoggingNestedError(t *testing.T) {
	err1 := goerr.New("e1", goerr.V("color", "orange"))
	err2 := goerr.Wrap(err1, "e2", goerr.V("number", "five"))
	out := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(out, nil))
	logger.Error("fail", slog.Any("error", err2))
	if !strings.Contains(out.String(), `"number":"five"`) {
		t.Errorf("Expected log output to contain '\"number\":\"five\"', got '%s'", out.String())
	}
	if !strings.Contains(out.String(), `"color":"orange"`) {
		t.Errorf("Expected log output to contain '\"color\":\"orange\"', got '%s'", out.String())
	}
}

func TestLoggerWithNil(t *testing.T) {
	out := &bytes.Buffer{}
	var err *goerr.Error
	logger := slog.New(slog.NewJSONHandler(out, nil))
	logger.Error("fail", slog.Any("error", err))
	if !strings.Contains(out.String(), `"error":null`) {
		t.Errorf("Expected log output to contain '\"error\":null', got '%s'", out.String())
	}
}

func TestUnstack(t *testing.T) {
	t.Run("original stack", func(t *testing.T) {
		err := oops()
		st := err.Stacks()
		if st == nil {
			t.Error("Expected stack trace to be nil")
		}
		if len(st) == 0 {
			t.Error("Expected stack trace length to be 0")
		}
		if st[0].Func != "github.com/m-mizutani/goerr/v2_test.oops" {
			t.Errorf("Not expected stack trace func name (github.com/m-mizutani/goerr/v2_test.oops): %s", st[0].Func)
		}
	})

	t.Run("unstacked", func(t *testing.T) {
		err := oops().Unstack()
		st1 := err.Stacks()
		if st1 == nil {
			t.Error("Expected stack trace to be non-nil")
		}
		if len(st1) == 0 {
			t.Error("Expected stack trace length to be non-zero")
		}
		if st1[0].Func != "github.com/m-mizutani/goerr/v2_test.TestUnstack.func2" {
			t.Errorf("Not expected stack trace func name (github.com/m-mizutani/goerr/v2_test.TestUnstack.func2): %s", st1[0].Func)
		}
	})

	t.Run("unstackN with 2", func(t *testing.T) {
		err := oops().UnstackN(2)
		st2 := err.Stacks()
		if st2 == nil {
			t.Error("Expected stack trace to be non-nil")
		}
		if len(st2) == 0 {
			t.Error("Expected stack trace length to be non-zero")
		}
		if st2[0].Func != "testing.tRunner" {
			t.Errorf("Not expected stack trace func name (testing.tRunner): %s", st2[0].Func)
		}
	})
}

func sliceHas(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

func TestTags(t *testing.T) {
	t1 := goerr.NewTag("tag1")
	t2 := goerr.NewTag("tag2")

	err1 := goerr.New("omg").WithTags(t1)
	err2 := fmt.Errorf("oops: %w", err1)
	err3 := goerr.Wrap(err2, "orange").WithTags(t2)
	err4 := fmt.Errorf("oh no: %w", err3)

	tags := goerr.Tags(err4)
	if len(tags) != 2 {
		t.Errorf("Expected tags length to be 2, got %d", len(tags))
	}
	if !sliceHas(tags, "tag1") {
		t.Error("Expected tags to contain 'tag1'")
	}
	if !sliceHas(tags, "tag2") {
		t.Error("Expected tags to contain 'tag2'")
	}
}

func TestValues(t *testing.T) {
	err1 := goerr.New("omg", goerr.V("color", "blue"))
	err2 := fmt.Errorf("oops: %w", err1)
	err3 := goerr.Wrap(err2, "red", goerr.V("number", "five"))
	err4 := fmt.Errorf("oh no: %w", err3)

	values := goerr.Values(err4)
	if len(values) != 2 {
		t.Errorf("Expected values length to be 2, got %d", len(values))
	}
	if values["color"] != "blue" {
		t.Errorf("Expected value for 'color' to be 'blue', got '%v'", values["color"])
	}
	if values["number"] != "five" {
		t.Errorf("Expected value for 'number' to be 'five', got '%v'", values["number"])
	}
}

func TestFormat(t *testing.T) {
	err := goerr.New("omg", goerr.V("color", "blue"), goerr.V("number", 123))

	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%+v", err)
	if !strings.Contains(b.String(), "color: blue") {
		t.Errorf("Expected log output to contain 'color: blue', got '%s'", b.String())
	}
	if !strings.Contains(b.String(), "number: 123") {
		t.Errorf("Expected log output to contain 'number: 123', got '%s'", b.String())
	}
}

// Helper functions for MarshalJSON tests
func createSimpleError() *goerr.Error {
	return goerr.New("simple error message")
}

func createErrorWithValues() *goerr.Error {
	return goerr.New("error with values",
		goerr.Value("key1", "value1"),
		goerr.Value("key2", 42),
		goerr.Value("key3", true))
}

func createErrorWithTags() *goerr.Error {
	tag1 := goerr.NewTag("tag1")
	tag2 := goerr.NewTag("tag2")
	return goerr.New("error with tags",
		goerr.Tag(tag1),
		goerr.Tag(tag2))
}

func createWrappedErrorChain() *goerr.Error {
	inner := goerr.New("inner error", goerr.Value("inner_key", "inner_value"))
	return goerr.Wrap(inner, "outer error", goerr.Value("outer_key", "outer_value"))
}

func createMixedErrorChain() *goerr.Error {
	stdErr := errors.New("standard error")
	return goerr.Wrap(stdErr, "wrapped standard error", goerr.Value("wrapper_key", "wrapper_value"))
}

func setupComplexError() *goerr.Error {
	tag1 := goerr.NewTag("critical")
	tag2 := goerr.NewTag("system")
	inner := goerr.New("database connection failed", goerr.Value("db_host", "localhost"))
	return goerr.Wrap(inner, "service unavailable",
		goerr.Value("service", "user-service"),
		goerr.Value("timestamp", "2024-01-01T12:00:00Z"),
		goerr.Tag(tag1),
		goerr.Tag(tag2))
}

func TestError_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		setupErr func() *goerr.Error
		validate func(t *testing.T, jsonBytes []byte)
	}{
		{
			name:     "simple error",
			setupErr: createSimpleError,
			validate: func(t *testing.T, jsonBytes []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(jsonBytes, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				if result["message"] != "simple error message" {
					t.Errorf("Expected message 'simple error message', got %v", result["message"])
				}
				if result["id"] == nil || result["id"] == "" {
					t.Error("Expected non-empty id")
				}
				if result["stacktrace"] == nil {
					t.Error("Expected stacktrace to be present")
				}
			},
		},
		{
			name:     "error with values",
			setupErr: createErrorWithValues,
			validate: func(t *testing.T, jsonBytes []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(jsonBytes, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				values, ok := result["values"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected values to be object")
				}
				if values["key1"] != "value1" {
					t.Errorf("Expected key1='value1', got %v", values["key1"])
				}
				if values["key2"] != float64(42) { // JSON numbers are float64
					t.Errorf("Expected key2=42, got %v", values["key2"])
				}
				if values["key3"] != true {
					t.Errorf("Expected key3=true, got %v", values["key3"])
				}
			},
		},
		{
			name:     "error with tags",
			setupErr: createErrorWithTags,
			validate: func(t *testing.T, jsonBytes []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(jsonBytes, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				tags, ok := result["tags"].([]interface{})
				if !ok {
					t.Fatal("Expected tags to be array")
				}
				if len(tags) != 2 {
					t.Errorf("Expected 2 tags, got %d", len(tags))
				}
				tagSet := make(map[string]bool)
				for _, tag := range tags {
					tagSet[tag.(string)] = true
				}
				if !tagSet["tag1"] || !tagSet["tag2"] {
					t.Errorf("Expected tags 'tag1' and 'tag2', got %v", tags)
				}
			},
		},
		{
			name:     "wrapped error chain",
			setupErr: createWrappedErrorChain,
			validate: func(t *testing.T, jsonBytes []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(jsonBytes, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				if result["message"] != "outer error" {
					t.Errorf("Expected outer message, got %v", result["message"])
				}
				cause, ok := result["cause"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected cause to be object")
				}
				if cause["message"] != "inner error" {
					t.Errorf("Expected inner message, got %v", cause["message"])
				}
				// Check merged values
				values, ok := result["values"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected values to be object")
				}
				if values["inner_key"] != "inner_value" {
					t.Errorf("Expected inner_key='inner_value', got %v", values["inner_key"])
				}
				if values["outer_key"] != "outer_value" {
					t.Errorf("Expected outer_key='outer_value', got %v", values["outer_key"])
				}
			},
		},
		{
			name:     "mixed error chain",
			setupErr: createMixedErrorChain,
			validate: func(t *testing.T, jsonBytes []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(jsonBytes, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				cause, ok := result["cause"].(string)
				if !ok {
					t.Fatal("Expected cause to be string for standard error")
				}
				if cause != "standard error" {
					t.Errorf("Expected cause='standard error', got %v", cause)
				}
			},
		},
		{
			name:     "nil error",
			setupErr: func() *goerr.Error { return nil },
			validate: func(t *testing.T, jsonBytes []byte) {
				expected := "null"
				if string(jsonBytes) != expected {
					t.Errorf("Expected %s, got %s", expected, string(jsonBytes))
				}
			},
		},
		{
			name:     "complex error",
			setupErr: setupComplexError,
			validate: func(t *testing.T, jsonBytes []byte) {
				var result map[string]interface{}
				if err := json.Unmarshal(jsonBytes, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}
				// Verify all components
				if result["message"] == nil {
					t.Error("Expected message")
				}
				if result["values"] == nil {
					t.Error("Expected values")
				}
				if result["tags"] == nil {
					t.Error("Expected tags")
				}
				if result["cause"] == nil {
					t.Error("Expected cause")
				}
				if result["stacktrace"] == nil {
					t.Error("Expected stacktrace")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable for Go versions < 1.22
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setupErr()
			jsonBytes, marshalErr := json.Marshal(err)
			if marshalErr != nil {
				t.Fatalf("Failed to marshal error: %v", marshalErr)
			}
			tt.validate(t, jsonBytes)
		})
	}
}

func TestError_MarshalJSON_Integration(t *testing.T) {
	err := createComplexErrorForIntegration()

	// Test with json.Marshal
	jsonBytes, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Fatalf("json.Marshal failed: %v", marshalErr)
	}

	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(jsonBytes, &result); unmarshalErr != nil {
		t.Fatalf("json.Unmarshal failed: %v", unmarshalErr)
	}

	if result["message"] == nil {
		t.Error("Expected message in JSON output")
	}
}

func TestError_MarshalJSON_Encoder(t *testing.T) {
	err := createSimpleError()

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if encodeErr := encoder.Encode(err); encodeErr != nil {
		t.Fatalf("json.NewEncoder().Encode() failed: %v", encodeErr)
	}

	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(buf.Bytes(), &result); unmarshalErr != nil {
		t.Fatalf("json.Unmarshal failed: %v", unmarshalErr)
	}

	if result["message"] != "simple error message" {
		t.Errorf("Expected message 'simple error message', got %v", result["message"])
	}
}

func TestError_MarshalJSON_Compatibility(t *testing.T) {
	err := createComplexErrorForIntegration()

	// Marshal using MarshalJSON
	jsonBytes1, err1 := json.Marshal(err)
	if err1 != nil {
		t.Fatalf("json.Marshal(err) failed: %v", err1)
	}

	// Marshal using Printable
	jsonBytes2, err2 := json.Marshal(err.Printable())
	if err2 != nil {
		t.Fatalf("json.Marshal(err.Printable()) failed: %v", err2)
	}

	// Unmarshal both to compare structure instead of string
	var result1, result2 map[string]interface{}
	if err := json.Unmarshal(jsonBytes1, &result1); err != nil {
		t.Fatalf("Failed to unmarshal MarshalJSON result: %v", err)
	}
	if err := json.Unmarshal(jsonBytes2, &result2); err != nil {
		t.Fatalf("Failed to unmarshal Printable result: %v", err)
	}
	
	// Normalize tags order since map iteration is non-deterministic
	normalizeTagsOrder(result1)
	normalizeTagsOrder(result2)
	
	// Compare the structures
	if !reflect.DeepEqual(result1, result2) {
		t.Error("MarshalJSON output should match Printable() output")
		t.Logf("MarshalJSON: %+v", result1)
		t.Logf("Printable:   %+v", result2)
	}
}

// normalizeTagsOrder sorts tags arrays in the JSON structure recursively
// to ensure deterministic comparison
func normalizeTagsOrder(data map[string]interface{}) {
	if tags, ok := data["tags"].([]interface{}); ok {
		// Convert to string slice, sort, and convert back
		tagStrings := make([]string, len(tags))
		for i, tag := range tags {
			tagStrings[i] = tag.(string)
		}
		sort.Strings(tagStrings)
		
		// Convert back to []interface{}
		sortedTags := make([]interface{}, len(tagStrings))
		for i, tag := range tagStrings {
			sortedTags[i] = tag
		}
		data["tags"] = sortedTags
	}
	
	// Recursively normalize nested cause objects
	if cause, ok := data["cause"].(map[string]interface{}); ok {
		normalizeTagsOrder(cause)
	}
}

func createComplexErrorForIntegration() *goerr.Error {
	tag := goerr.NewTag("integration")
	inner := goerr.New("inner error", goerr.Value("inner", "value"))
	return goerr.Wrap(inner, "integration test error",
		goerr.Value("test", "integration"),
		goerr.Tag(tag))
}

func BenchmarkError_MarshalJSON(b *testing.B) {
	err := setupComplexError()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(err)
	}
}

func BenchmarkError_MarshalJSON_vs_Printable(b *testing.B) {
	err := setupComplexError()

	b.Run("MarshalJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(err)
		}
	})

	b.Run("Printable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(err.Printable())
		}
	})
}
