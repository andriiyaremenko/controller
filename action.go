package controller

import (
	"context"
	"net/http"
)

// Action is http.Handle that utilizes generics to process request payload
// and reduce amount of boilerplate code.
type Action[T, U any] func(context.Context, T) (U, error)

// With allows change default Action behaviour with options.
func (handle Action[T, U]) With(opts ...func(*ActionOptions)) http.Handler {
	options := ActionOptions{
		Options: Options{
			LogError:         func(context.Context, error, string) {},
			ErrorHandlers:    []ErrorHandler{},
			ReadRequestParam: map[string]func(*http.Request, string) string{},
			WriteResponse:    JSONWriter,
			SuccessCode:      http.StatusOK,
		},
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
			&ActionOptions{
				Options: Options{
					LogError:         func(context.Context, error, string) {},
					ErrorHandlers:    []ErrorHandler{},
					ReadRequestParam: map[string]func(*http.Request, string) string{},
					WriteResponse:    JSONWriter,
					SuccessCode:      http.StatusOK,
				},
				ReadRequestContent: JSONBodyReader,
			},
		).
		ServeHTTP(w, r)
}

func (handle Action[T, U]) getHttpHandle(opts *ActionOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := map[string]string{}

		for key, fn := range opts.ReadRequestParam {
			params[key] = fn(r, key)
		}

		ctx = setContextParam(ctx, params)

		var model T
		if err := opts.ReadRequestContent(r, &model); err != nil {
			code, response := getErrorResponse(ctx, err, opts.ErrorHandlers)

			opts.LogError(ctx, err, "failed to read request content")
			opts.WriteResponse(ctx, w, opts.LogError, code, response)

			return
		}

		result, err := handle(ctx, model)
		if err != nil {
			code, response := getErrorResponse(ctx, err, opts.ErrorHandlers)

			opts.LogError(ctx, err, "request failed")
			opts.WriteResponse(ctx, w, opts.LogError, code, response)

			return
		}

		opts.WriteResponse(ctx, w, opts.LogError, opts.SuccessCode, result)
	}
}
