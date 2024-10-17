package controller

import (
	"net/http"
	"runtime/debug"
)

// Handle is http.Handle that utilizes generics to process request payload
// and reduce amount of boilerplate code.
type Handle[T any] func(*http.Request) (T, error)

// With allows change default Action behaviour with options.
func (handle Handle[T]) With(opts ...func(Options)) http.Handler {
	options := options{successCode: http.StatusOK, writeResponse: WriteJSON}
	for _, option := range opts {
		option(&options)
	}

	return handle.getHttpHandle(&options)
}

func (handle Handle[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handle.
		getHttpHandle(&options{successCode: http.StatusOK, writeResponse: WriteJSON}).
		ServeHTTP(w, r)
}

func (handle Handle[T]) getHttpHandle(opts *options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				stack := debug.Stack()
				err := newRecoveredError(rp, stack)

				logger().ErrorContext(ctx, "request failed: recovered from panic during request", "error", err, "stack", string(stack))

				response, code := getErrorResponse(r, err, opts.errorHandlers)
				opts.writeResponse(ctx, w, response, code)

			}
		}()

		result, err := handle(r)
		if err != nil {
			response, code := getErrorResponse(r, err, opts.errorHandlers)

			logger().ErrorContext(ctx, "request failed", "error", err)
			opts.writeResponse(ctx, w, response, code)

			return
		}

		opts.writeResponse(ctx, w, result, opts.successCode)
	}
}
