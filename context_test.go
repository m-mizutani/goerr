package goerr_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/goerr"
)

func funcWithContext(ctx context.Context, color string) error {
	if color != "red" {
		return goerr.New("color is not red").With("color", color).WithContext(ctx)
	}
	return nil
}

func TestContext(t *testing.T) {
	reqID := "req-123"

	ctx := context.Background()
	ctx = goerr.InjectValue(ctx, "request_id", reqID)

	err := funcWithContext(ctx, "blue")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	goErr := goerr.Unwrap(err)
	if goErr == nil {
		t.Error("Expected goerr.Error, got other type")
	}

	if goErr.Values()["color"] != "blue" {
		t.Errorf("Expected color value to be 'blue', got '%v'", goErr.Values()["color"])
	}

	if goErr.Values()["request_id"] != reqID {
		t.Errorf("Expected request_id value to be '%s', got '%v'", reqID, goErr.Values()["request_id"])
	}
}
