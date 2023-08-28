package controller

import (
	"context"
	"net/http"
)

// Action is http.Handle that utilizes generics to process request payload
// and reduce amount of boilerplate code.
type Action[T, U any] func(context.Context, T) (U, error)

// With allows change default Action behaviour with options.
func (handle Action[T, U]) With(opts ...func(Options)) http.Handler {
	options := options{
		logError:           func(context.Context, error, string) {},
		errorHandlers:      []ErrorHandler{},
		readRequestParam:   map[string]func(*http.Request, string) string{},
		writeResponse:      JSONWriter,
		successCode:        http.StatusOK,
		readRequestContent: JSONBodyReader,
	}
	for _, option := range opts {
		option(&options)
	}

	return handle.getHttpHandle(&options)
}

func (handle Action[T, U]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handle.
		getHttpHandle(
			&options{
				logError:           func(context.Context, error, string) {},
				errorHandlers:      []ErrorHandler{},
				readRequestParam:   map[string]func(*http.Request, string) string{},
				writeResponse:      JSONWriter,
				successCode:        http.StatusOK,
				readRequestContent: JSONBodyReader,
			},
		).
		ServeHTTP(w, r)
}

func (handle Action[T, U]) getHttpHandle(opts *options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := map[string]string{}

		for key, fn := range opts.readRequestParam {
			params[key] = fn(r, key)
		}

		ctx = setContextParam(ctx, params)

		var model T
		if err := opts.readRequestContent(r, &model); err != nil {
			code, response := getErrorResponse(ctx, err, opts.errorHandlers)

			opts.logError(ctx, err, "failed to read request content")
			opts.writeResponse(ctx, w, opts.logError, code, response)

			return
		}

		result, err := handle(ctx, model)
		if err != nil {
			code, response := getErrorResponse(ctx, err, opts.errorHandlers)

			opts.logError(ctx, err, "request failed")
			opts.writeResponse(ctx, w, opts.logError, code, response)

			return
		}

		opts.writeResponse(ctx, w, opts.logError, opts.successCode, result)
	}
}
