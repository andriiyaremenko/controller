package controller

import (
	"context"
	"errors"
	"net/http"
)

type Options struct {
	ReadRequestContent func(*http.Request, any) error
	RequestURLParam    func(*http.Request, string) string
	LogError           func(context.Context, error, string)
	WriteResponse      func(
		context.Context, http.ResponseWriter,
		func(context.Context, error, string),
		int, any,
	)

	ErrorHandlers []ErrorHandler
	SuccessCode   int
}

func WithHTTPStatus(code int) func(*Options) {
	return func(opts *Options) {
		opts.SuccessCode = code
	}
}

func WithURLParamReader(r func(*http.Request, string) string) func(*Options) {
	return func(opts *Options) {
		opts.RequestURLParam = r
	}
}

func WithErrorLogger(log func(context.Context, error, string)) func(*Options) {
	return func(opts *Options) {
		opts.LogError = log
	}
}

func WithAppError[E any](httpCode int) func(*Options) {
	return func(opts *Options) {
		opts.ErrorHandlers = append(
			opts.ErrorHandlers,
			func(err error) (int, any) {
				var errModel E
				if errors.As(err, &errModel) {
					return httpCode, errModel
				}

				return 0, nil
			},
		)
	}
}

func WithResponseWriter(w ResponseWriter) func(*Options) {
	return func(opts *Options) {
		opts.WriteResponse = w
	}
}

func WithRequestContentReader(r RequestReader) func(*Options) {
	return func(opts *Options) {
		opts.ReadRequestContent = r
	}
}
