package controller

import (
	"context"
	"net/http"
)

func Task[T any](
	handle func(context.Context, func(ParamSource, string) string) (T, error),
	options ...func(*Options),
) http.HandlerFunc {
	opts := Options{
		LogError:        func(context.Context, error, string) {},
		ErrorHandlers:   []ErrorHandler{},
		RequestURLParam: func(*http.Request, string) string { return "" },
		WriteResponse:   JSONWriter,
		SuccessCode:     http.StatusOK,
	}
	for _, option := range options {
		option(&opts)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		result, err := handle(ctx, readParam(req, opts.RequestURLParam))
		if err != nil {
			code, response := getErrorResponse(err, opts.ErrorHandlers)

			opts.LogError(ctx, err, "request failed")
			opts.WriteResponse(ctx, w, opts.LogError, code, response)

			return
		}

		opts.WriteResponse(ctx, w, opts.LogError, opts.SuccessCode, result)
	}
}
