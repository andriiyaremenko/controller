package controller

import (
	"context"
	"net/http"
)

func Action[T, U any](
	handle func(context.Context, T, func(ParamSource, string) string) (U, error),
	options ...func(*Options),
) http.HandlerFunc {
	opts := Options{
		LogError:           func(context.Context, error, string) {},
		ErrorHandlers:      []ErrorHandler{},
		RequestURLParam:    func(*http.Request, string) string { return "" },
		WriteResponse:      JSONWriter,
		SuccessCode:        http.StatusOK,
		ReadRequestContent: JSONBodyReader,
	}
	for _, option := range options {
		option(&opts)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		var model T
		if err := opts.ReadRequestContent(req, &model); err != nil {
			code, response := getErrorResponse(err, opts.ErrorHandlers)

			opts.LogError(ctx, err, "failed to read request content")
			opts.WriteResponse(ctx, w, opts.LogError, code, response)

			return
		}

		result, err := handle(ctx, model, readParam(req, opts.RequestURLParam))
		if err != nil {
			code, response := getErrorResponse(err, opts.ErrorHandlers)

			opts.LogError(ctx, err, "request failed")
			opts.WriteResponse(ctx, w, opts.LogError, code, response)

			return
		}

		opts.WriteResponse(ctx, w, opts.LogError, opts.SuccessCode, result)
	}
}
