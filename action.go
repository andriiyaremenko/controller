package controller

import (
	"context"
	"net/http"
)

type Action[T, U any] func(context.Context, T, func(ParamSource, string) string) (U, error)

func (handle Action[T, U]) With(opts ...func(*Options)) http.Handler {
	options := Options{
		LogError:           func(context.Context, error, string) {},
		ErrorHandlers:      []ErrorHandler{},
		RequestURLParam:    func(*http.Request, string) string { return "" },
		WriteResponse:      JSONWriter,
		SuccessCode:        http.StatusOK,
		ReadRequestContent: JSONBodyReader,
	}
	for _, option := range opts {
		option(&options)
	}

	return handle.getHttpHandle(&options)
}

func (handle Action[T, U]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handle.
		getHttpHandle(
			&Options{
				LogError:           func(context.Context, error, string) {},
				ErrorHandlers:      []ErrorHandler{},
				RequestURLParam:    func(*http.Request, string) string { return "" },
				WriteResponse:      JSONWriter,
				SuccessCode:        http.StatusOK,
				ReadRequestContent: JSONBodyReader,
			},
		).
		ServeHTTP(w, r)
}

func (handle Action[T, U]) getHttpHandle(opts *Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var model T
		if err := opts.ReadRequestContent(r, &model); err != nil {
			code, response := getErrorResponse(err, opts.ErrorHandlers)

			opts.LogError(ctx, err, "failed to read request content")
			opts.WriteResponse(ctx, w, opts.LogError, code, response)

			return
		}

		result, err := handle(ctx, model, readParam(r, opts.RequestURLParam))
		if err != nil {
			code, response := getErrorResponse(err, opts.ErrorHandlers)

			opts.LogError(ctx, err, "request failed")
			opts.WriteResponse(ctx, w, opts.LogError, code, response)

			return
		}

		opts.WriteResponse(ctx, w, opts.LogError, opts.SuccessCode, result)
	}
}
