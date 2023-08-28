package controller

import (
	"context"
)

type key int

var paramKey key

func ContextParam(ctx context.Context, key string) string {
	params, ok := ctx.Value(paramKey).(map[string]string)
	if !ok {
		return ""
	}

	v, ok := params[key]
	if !ok {
		return ""
	}

	return v
}

func setContextParam(ctx context.Context, v map[string]string) context.Context {
	return context.WithValue(ctx, paramKey, v)
}
