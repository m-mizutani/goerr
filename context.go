package goerr

import "context"

type errContext struct {
	values map[string]any
}

type errContextKey struct{}

func InjectValue(ctx context.Context, key string, value any) context.Context {
	newCtx := errContext{
		values: make(map[string]any),
	}
	oldCtx, ok := ctx.Value(errContextKey{}).(*errContext)
	if ok {
		for k, v := range oldCtx.values {
			newCtx.values[k] = v
		}
	}

	newCtx.values[key] = value
	return context.WithValue(ctx, errContextKey{}, &newCtx)
}
