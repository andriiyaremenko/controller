package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// ErrorHandler transforms error to HTTP status code and response body.
type ErrorHandler func(context.Context, error) (int, any)

// IfError checks if error is of E type
// and returns designated HTTP Status Code with E instance as a response if true.
func IfError[E any](httpCode int) ErrorHandler {
	return func(_ context.Context, err error) (int, any) {
		var target E
		if errors.As(err, &target) {
			return httpCode, target
		}

		return 0, nil
	}
}

// IfErrorUse checks if error is of E type
// and returns designated HTTP Status Code with transformed response using as callback if true.
func IfErrorUse[E any](as func(context.Context, E) any, httpCode int) ErrorHandler {
	return func(ctx context.Context, err error) (int, any) {
		var target E
		if errors.As(err, &target) {
			return httpCode, as(ctx, target)
		}

		return 0, nil
	}
}

// If request payload reading failed - ReadRequestError is returned.
type ReadRequestError struct {
	err error
}

func (err *ReadRequestError) Error() string {
	return fmt.Sprintf("failed to read request: %s", err.err)
}

func (err *ReadRequestError) Unwrap() error {
	return err.err
}

func getErrorResponse(ctx context.Context, err error, handlers []ErrorHandler) (int, any) {
	for _, handle := range handlers {
		code, response := handle(ctx, err)
		if code != 0 {
			return code, response
		}
	}

	return http.StatusInternalServerError, err.Error()
}
