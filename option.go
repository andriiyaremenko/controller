package controller

import (
	"errors"
	"net/http"
)

type Options interface {
	SuccessCode(int)
	ErrorHandlers(...ErrorMatcher)
	WriteResponse(WriteResponse)
}

type options struct {
	responseWriter WriteResponse
	errorHandlers  []ErrorMatcher
	successCode    int
}

func (o *options) SuccessCode(code int) {
	o.successCode = code
}

func (o *options) ErrorHandlers(handlers ...ErrorMatcher) {
	o.errorHandlers = append(o.errorHandlers, handlers...)
}

func (o *options) WriteResponse(w WriteResponse) {
	o.responseWriter = w
}

// Sets success response HTTP Status Code.
func SuccessCode(code int) func(Options) {
	return func(o Options) { o.SuccessCode(code) }
}

// ErrorWithCode checks if error is of E type
// and returns designated HTTP Status Code with E instance as a response if true.
func ErrorWithCode[E any](httpCode int) func(Options) {
	return func(o Options) {
		o.ErrorHandlers(
			MatchError(func(err error) (any, int) {
				var target E
				if errors.As(err, &target) {
					return target, httpCode
				}

				return nil, 0
			}),
		)
	}
}

// ErrorHandle checks if error is of E type
// and returns designated HTTP Status Code with transformed response using as callback if true.
func ErrorHandle(matcher ErrorMatcher) func(Options) {
	return func(o Options) {
		o.ErrorHandlers(
			MatchErrorRequest(func(r *http.Request, err error) (any, int) {
				response, code := matcher.Match(r, err)
				if code != 0 {
					return response, code
				}

				return nil, 0
			}),
		)
	}
}

// Sets response writer.
func ResponseWriter(w WriteResponse) func(Options) {
	return func(o Options) { o.WriteResponse(w) }
}
