package controller

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
)

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

type RecoveredError struct {
	Panic any
	Stack []byte
}

func (err *RecoveredError) Unwrap() error {
	rp, ok := err.Panic.(error)
	if ok {
		return rp
	}

	return nil
}

func (err *RecoveredError) Error() string {
	return fmt.Sprintf("recovered from panic: %v", err.Panic)
}

type ErrorMatcher interface {
	Match(*http.Request, error) (any, int)
}

// MatchErrorRequest transforms error to HTTP status code and response body.
type MatchErrorRequest func(*http.Request, error) (any, int)

func (match MatchErrorRequest) Match(r *http.Request, err error) (any, int) {
	return match(r, err)
}

type MatchError func(error) (any, int)

func (match MatchError) Match(_ *http.Request, err error) (any, int) {
	return match(err)
}

func SetDefaultErrorHandlers(handlers ...ErrorMatcher) {
	if len(handlers) > 0 {
		handlers = append(handlers, readRequestErrorHandle)
		defaultErrorHandlers.Store(&handlers)
	}
}

var readRequestErrorHandle = MatchError(func(err error) (any, int) {
	var readErr *ReadRequestError
	if errors.As(err, &readErr) {
		return readErr.Error(), http.StatusBadRequest
	}

	return nil, 0
})

func init() {
	defaultErrorHandlers.Store(&[]ErrorMatcher{readRequestErrorHandle})
}

var defaultErrorHandlers atomic.Pointer[[]ErrorMatcher]

func getErrorResponse(r *http.Request, err error, handlers []ErrorMatcher) (any, int) {
	for _, matcher := range append(handlers, *defaultErrorHandlers.Load()...) {
		response, code := matcher.Match(r, err)
		if code != 0 {
			return response, code
		}
	}

	return err.Error(), http.StatusInternalServerError
}

func newRecoveredError(p any, stack []byte) error {
	return &RecoveredError{
		Panic: p,
		Stack: stack,
	}
}
