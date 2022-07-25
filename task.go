package controller

import (
	"context"
	"net/http"
)

type Task[T any] func(context.Context, func(ParamSource, string) string) (T, error)

func (handle Task[T]) With(opts ...func(*Options)) http.Handler {
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

func (handle Task[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (handle Task[T]) getHttpHandle(opts *Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		result, err := handle(ctx, readParam(r, opts.RequestURLParam))
		if err != nil {
			code, response := getErrorResponse(err, opts.ErrorHandlers)

			opts.LogError(ctx, err, "request failed")
			opts.WriteResponse(ctx, w, opts.LogError, code, response)

			return
		}

		opts.WriteResponse(ctx, w, opts.LogError, opts.SuccessCode, result)
	}
}
