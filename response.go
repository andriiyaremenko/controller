package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
)

type DecoratedResponse[T, U any] func(U) Respond[T]

func (decorated DecoratedResponse[T, U]) With(opts ...func(Options)) func(U) http.Handler {
	return func(u U) http.Handler {
		return decorated(u).With(opts...)
	}
}

// Respond is http.Handler that utilizes generics to process request payload
// and reduce amount of boilerplate code.
type Respond[T any] func(*http.Request) (T, error)

// With allows change default Respond behaviour with options.
func (handle Respond[T]) With(opts ...func(Options)) http.Handler {
	options := options{successCode: http.StatusOK, responseWriter: WriteJSON}
	for _, option := range opts {
		option(&options)
	}

	return handle.getHttpHandle(&options)
}

func (handle Respond[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handle.
		getHttpHandle(&options{successCode: http.StatusOK, responseWriter: WriteJSON}).
		ServeHTTP(w, r)
}

func (handle Respond[T]) getHttpHandle(opts *options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				stack := debug.Stack()
				err := newRecoveredError(rp, stack)

				logger().Error("request failed: recovered from panic during request", "error", err, "stack", string(stack))

				response, code := getErrorResponse(r, err, opts.errorHandlers)
				opts.responseWriter.WriteError(ctx, w, response, code)
			}
		}()

		result, err := handle(r)
		if err != nil {
			logger().Error("request failed", "error", err)

			response, code := getErrorResponse(r, err, opts.errorHandlers)
			opts.responseWriter.WriteError(ctx, w, response, code)

			return
		}

		opts.responseWriter.Write(ctx, w, result, opts.successCode)
	}
}

type WriteResponse interface {
	Write(context.Context, http.ResponseWriter, any, int)
	WriteError(context.Context, http.ResponseWriter, any, int)
}

// Response writer type.
type WriteResponseFn func(context.Context, http.ResponseWriter, any, int)

func (fn WriteResponseFn) Write(ctx context.Context, w http.ResponseWriter, value any, code int) {
	fn(ctx, w, value, code)
}

func (fn WriteResponseFn) WriteError(ctx context.Context, w http.ResponseWriter, err any, code int) {
	fn(ctx, w, err, code)
}

// Response writer to write JSON response
// in body with Content-Type "application/json; charset=utf-8" Header.
var WriteJSON WriteResponseFn = func(ctx context.Context, w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger().Error("failed to write JSON", "error", err)
	}
}
