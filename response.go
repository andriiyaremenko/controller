package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type ResponseWriter func(
	context.Context, http.ResponseWriter,
	func(context.Context, error, string),
	int, any,
)

func NoContentWriter(
	_ context.Context, w http.ResponseWriter,
	_ func(context.Context, error, string),
	_ int, _ any,
) {
	w.WriteHeader(http.StatusNoContent)
}

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

func FormWriter(encode func(any, url.Values) error) ResponseWriter {
	return func(
		ctx context.Context, w http.ResponseWriter,
		logError func(context.Context, error, string),
		status int, data any,
	) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.WriteHeader(status)

		if data == nil {
			return
		}

		form := url.Values{}
		if err := encode(data, form); err != nil {
			logError(ctx, err, "failed to write Form Data")
		}

		if _, err := w.Write([]byte(form.Encode())); err != nil {
			logError(ctx, err, "failed to write Form Data")
		}
	}
}
