package controller

import (
	"context"
	"net/http"
)

type Options interface {
	SetSuccessCode(int)
	SetReadRequestParam(string, func(*http.Request, string) string)
	SetErrorLogger(func(context.Context, error, string))
	SetErrorHandlers(...ErrorHandler)
	SetWriteResponse(WriteResponse)
	SetReadRequestContent(ReadRequest)
}

// Shared Action and Task options.
type options struct {
	readRequestParam map[string]func(*http.Request, string) string
	logError         func(context.Context, error, string)
	writeResponse    func(
		context.Context, http.ResponseWriter,
		func(context.Context, error, string),
		int, any,
	)

	errorHandlers      []ErrorHandler
	successCode        int
	readRequestContent ReadRequest
}

func (o *options) SetSuccessCode(code int) {
	o.successCode = code
}

func (o *options) SetReadRequestParam(key string, fn func(*http.Request, string) string) {
	o.readRequestParam[key] = fn
}

func (o *options) SetErrorLogger(log func(context.Context, error, string)) {
	o.logError = log
}

func (o *options) SetErrorHandlers(handlers ...ErrorHandler) {
	o.errorHandlers = append(o.errorHandlers, handlers...)
}

func (o *options) SetWriteResponse(w WriteResponse) {
	o.writeResponse = w
}

func (o *options) SetReadRequestContent(r ReadRequest) {
	o.readRequestContent = r
}

// Sets success response HTTP Status Code.
func HTTPStatus(code int) func(Options) {
	return func(o Options) { o.SetSuccessCode(code) }
}

// Sets URL parameters reader.
func RequestParam(key string, fn func(*http.Request, string) string) func(Options) {
	return func(o Options) { o.SetReadRequestParam(key, fn) }
}

// Sets logger to log error results.
func ErrorLogger(log func(context.Context, error, string)) func(Options) {
	return func(o Options) { o.SetErrorLogger(log) }
}

// Sets error handlers to return specific to each error HTTP Status Codes.
func ErrorHandlers(handlers ...ErrorHandler) func(Options) {
	return func(o Options) {
		o.SetErrorHandlers(handlers...)
	}
}

// Sets response writer.
func ResponseWriter(w WriteResponse) func(Options) {
	return func(o Options) { o.SetWriteResponse(w) }
}

// Sets requests content reader.
func RequestContentReader(r ReadRequest) func(Options) {
	return func(o Options) { o.SetReadRequestContent(r) }
}
