package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
)

// Respond is http.Handler that utilizes generics to process request payload
// and reduce amount of boilerplate code.
type Respond[T any] func(*http.Request) (T, error)

// With allows change default Respond behaviour with options.
func (handle Respond[T]) With(opts ...func(Options)) http.Handler {
	options := options{successCode: http.StatusOK, writeResponse: WriteJSON}
	for _, option := range opts {
		option(&options)
	}

	return handle.getHttpHandle(&options)
}

func (handle Respond[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handle.
		getHttpHandle(&options{successCode: http.StatusOK, writeResponse: WriteJSON}).
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
				opts.writeResponse(ctx, w, response, code)

			}
		}()

		result, err := handle(r)
		if err != nil {
			response, code := getErrorResponse(r, err, opts.errorHandlers)

			logger().Error("request failed", "error", err)
			opts.writeResponse(ctx, w, response, code)

			return
		}

		opts.writeResponse(ctx, w, result, opts.successCode)
	}
}

// Response writer type.
type WriteResponse func(context.Context, http.ResponseWriter, any, int)

// Response writer to write JSON response
// in body with Content-Type "application/json; charset=utf-8" Header.
func WriteJSON(ctx context.Context, w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger().Error("failed to write JSON", "error", err)
	}
}
