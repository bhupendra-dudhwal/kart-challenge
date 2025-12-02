package utils

import (
	"context"
	"fmt"
	"kart-challenge/internal/constants"
	"os"

	"github.com/valyala/fasthttp"
)

func FileExists(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("path exists but is a directory: %s", path)
		}
		return nil
	}

	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	return fmt.Errorf("failed to check file: %w", err)
}

func CtxValue[T any](ctx context.Context, key constants.CtxKey) (T, bool) {
	v := ctx.Value(key)
	return ctxValue[T](v)
}

func PathParamValue[T any](ctx *fasthttp.RequestCtx, key string) (T, bool) {
	v := ctx.UserValue(key)
	return ctxValue[T](v)
}

func HeaderValue[T any](ctx *fasthttp.RequestCtx, key string) (T, bool) {
	raw := ctx.Request.Header.Peek(key)
	if raw == nil {
		var zero T
		return zero, false
	}
	return convert[T](raw)
}

func ctxValue[T any](v any) (T, bool) {
	if v == nil {
		var zero T
		return zero, false
	}

	value, ok := v.(T)
	return value, ok
}

func convert[T any](v []byte) (T, bool) {
	var zero T

	switch any(zero).(type) {
	case string:
		return any(string(v)).(T), true

	case []byte:
		return any(v).(T), true

	default:
		return zero, false
	}
}
