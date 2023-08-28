package controller

import (
	"context"
	"net/http"
)

// Task is http.Handle that utilizes generics to reduce amount of boilerplate code.
type Task[T any] func(context.Context) (T, error)

// With allows change default Action behaviour with options.
func (handle Task[T]) With(opts ...func(*TaskOptions)) http.Handler {
	options := TaskOptions{
		Options: Options{
			LogError:         func(context.Context, error, string) {},
			ErrorHandlers:    []ErrorHandler{},
			ReadRequestParam: map[string]func(*http.Request, string) string{},
			WriteResponse:    JSONWriter,
			SuccessCode:      http.StatusOK,
		},
	}
	for _, option := range opts {
		option(&options)
	}

	return handle.getHttpHandle(&options)
}

func (handle Task[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handle.
		getHttpHandle(
			&TaskOptions{
				Options: Options{
					LogError:         func(context.Context, error, string) {},
					ErrorHandlers:    []ErrorHandler{},
					ReadRequestParam: map[string]func(*http.Request, string) string{},
					WriteResponse:    JSONWriter,
					SuccessCode:      http.StatusOK,
				},
			},
		).
		ServeHTTP(w, r)
}

func (handle Task[T]) getHttpHandle(opts *TaskOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := map[string]string{}

		for key, fn := range opts.ReadRequestParam {
			params[key] = fn(r, key)
		}

		ctx = setContextParam(ctx, params)
		result, err := handle(ctx)
		if err != nil {
			code, response := getErrorResponse(ctx, err, opts.ErrorHandlers)

			opts.LogError(ctx, err, "request failed")
			opts.WriteResponse(ctx, w, opts.LogError, code, response)

			return
		}

		opts.WriteResponse(ctx, w, opts.LogError, opts.SuccessCode, result)
	}
}
