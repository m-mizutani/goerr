package main

import (
	"context"
	"log/slog"

	"github.com/m-mizutani/goerr"
)

type requestID string

func someAction(id requestID) error {
	ctx := context.Background()
	ctx = goerr.InjectValue(ctx, "request_id", id)

	// ... some action

	return goerr.New("failed").WithContext(ctx)
}

func main() {
	id := requestID("req-123")
	if err := someAction(id); err != nil {
		slog.Default().Error("aborted", slog.Any("error", err))
	}
}
