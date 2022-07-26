package controller

import (
	"context"
	"net/http"
)

type ActionOptions struct {
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

func ActionHTTPStatus(code int) func(*ActionOptions) {
	return func(opts *ActionOptions) { opts.SuccessCode = code }
}

func ActionURLParamReader(r func(*http.Request, string) string) func(*ActionOptions) {
	return func(opts *ActionOptions) { opts.RequestURLParam = r }
}

func ActionErrorLogger(log func(context.Context, error, string)) func(*ActionOptions) {
	return func(opts *ActionOptions) { opts.LogError = log }
}

func ActionAppError(handler ErrorHandler) func(*ActionOptions) {
	return func(opts *ActionOptions) { opts.ErrorHandlers = append(opts.ErrorHandlers, handler) }
}

func ActionResponseWriter(w ResponseWriter) func(*ActionOptions) {
	return func(opts *ActionOptions) { opts.WriteResponse = w }
}

func ActionRequestContentReader(r RequestReader) func(*ActionOptions) {
	return func(opts *ActionOptions) { opts.ReadRequestContent = r }
}

type TaskOptions struct {
	RequestURLParam func(*http.Request, string) string
	LogError        func(context.Context, error, string)
	WriteResponse   func(
		context.Context, http.ResponseWriter,
		func(context.Context, error, string),
		int, any,
	)

	ErrorHandlers []ErrorHandler
	SuccessCode   int
}

func TaskHTTPStatus(code int) func(*TaskOptions) {
	return func(opts *TaskOptions) { opts.SuccessCode = code }
}

func TaskURLParamReader(r func(*http.Request, string) string) func(*TaskOptions) {
	return func(opts *TaskOptions) { opts.RequestURLParam = r }
}

func TaskErrorLogger(log func(context.Context, error, string)) func(*TaskOptions) {
	return func(opts *TaskOptions) { opts.LogError = log }
}

func TaskAppError(handler ErrorHandler) func(*TaskOptions) {
	return func(opts *TaskOptions) { opts.ErrorHandlers = append(opts.ErrorHandlers, handler) }
}

func TaskResponseWriter(w ResponseWriter) func(*TaskOptions) {
	return func(opts *TaskOptions) { opts.WriteResponse = w }
}
