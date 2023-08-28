package controller

import (
	"context"
	"net/http"
)

// Task is http.Handle that utilizes generics to reduce amount of boilerplate code.
type Task[T any] func(context.Context) (T, error)

// With allows change default Action behaviour with options.
func (handle Task[T]) With(opts ...func(Options)) http.Handler {
	options := options{
		logError:         func(context.Context, error, string) {},
		errorHandlers:    []ErrorHandler{},
		readRequestParam: map[string]func(*http.Request, string) string{},
		writeResponse:    JSONWriter,
		successCode:      http.StatusOK,
	}
	for _, option := range opts {
		option(&options)
	}

	return handle.getHttpHandle(&options)
}

func (handle Task[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handle.
		getHttpHandle(
			&options{
				logError:         func(context.Context, error, string) {},
				errorHandlers:    []ErrorHandler{},
				readRequestParam: map[string]func(*http.Request, string) string{},
				writeResponse:    JSONWriter,
				successCode:      http.StatusOK,
			},
		).
		ServeHTTP(w, r)
}

func (handle Task[T]) getHttpHandle(opts *options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := map[string]string{}

		for key, fn := range opts.readRequestParam {
			params[key] = fn(r, key)
		}

		ctx = setContextParam(ctx, params)
		result, err := handle(ctx)
		if err != nil {
			code, response := getErrorResponse(ctx, err, opts.errorHandlers)

			opts.logError(ctx, err, "request failed")
			opts.writeResponse(ctx, w, opts.logError, code, response)

			return
		}

		opts.writeResponse(ctx, w, opts.logError, opts.successCode, result)
	}
}
