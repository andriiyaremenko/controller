package controller

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorHandler transforms error to HTTP status code and response body.
type ErrorHandler func(error, ReadParam) (int, any)

// IfError checks if error is of E type
// and returns designated HTTP Status Code with E instance as a response if true.
func IfError[E any](httpCode int) ErrorHandler {
	return func(err error, readParam ReadParam) (int, any) {
		var target E
		if errors.As(err, &target) {
			return httpCode, target
		}

		return 0, nil
	}
}

// IfErrorUse checks if error is of E type
// and returns designated HTTP Status Code with transformed response using as callback if true.
func IfErrorUse[E any](as func(E, ReadParam) any, httpCode int) ErrorHandler {
	return func(err error, readParam ReadParam) (int, any) {
		var target E
		if errors.As(err, &target) {
			return httpCode, as(target, readParam)
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

func getErrorResponse(err error, paramReader ReadParam, handlers []ErrorHandler) (int, any) {
	for _, handle := range handlers {
		code, response := handle(err, paramReader)
		if code != 0 {
			return code, response
		}
	}

	return http.StatusInternalServerError, err.Error()
}
