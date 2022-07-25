package controller

import (
	"context"
	"encoding/json"
	"net/http"
)

type ResponseWriter func(
	context.Context, http.ResponseWriter,
	func(context.Context, error, string),
	int, any,
)

func JSONWriter(
	ctx context.Context, w http.ResponseWriter,
	logError func(context.Context, error, string),
	status int, data any,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logError(ctx, err, "failed to write JSON")
	}
}
