package controller

import (
	"context"
	"encoding/json"
	"net/http"
)

// Response writer type.
type WriteResponse func(context.Context, http.ResponseWriter, any, int)

// Response writer to write JSON response
// in body with Content-Type "application/json; charset=utf-8" Header.
func WriteJSON(ctx context.Context, w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger().ErrorContext(ctx, "failed to write JSON", "error", err)
	}
}
