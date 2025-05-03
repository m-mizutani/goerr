package goerr_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
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
	if st[0].Line != 16 {
		t.Errorf("Expected line number 13, got %d", st[0].Line)
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
	if !strings.Contains(b.String(), "- color=blue") {
		t.Errorf("Expected log output to contain 'color=blue', got '%s'", b.String())
	}
}
