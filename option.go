package controller

import (
	"context"
	"net/http"
)

// Shared Action and Task options.
type Options struct {
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

func (o *Options) Set(update func(*Options)) {
	update(o)
}

//  Task options.
type TaskOptions struct {
	Options
}

//  Action options.
type ActionOptions struct {
	Options

	ReadRequestContent func(*http.Request, any) error
}

//  Task and Action Options union type.
type Option interface {
	*TaskOptions | *ActionOptions
	Set(func(*Options))
}

// Sets success response HTTP Status Code.
func HTTPStatus[O Option](code int) func(O) {
	return func(o O) { o.Set(func(opts *Options) { opts.SuccessCode = code }) }
}

// Sets URL parameters reader.
func URLParamReader[O Option](r func(*http.Request, string) string) func(O) {
	return func(o O) { o.Set(func(opts *Options) { opts.RequestURLParam = r }) }
}

// Sets logger to log error results.
func ErrorLogger[O Option](log func(context.Context, error, string)) func(O) {
	return func(o O) { o.Set(func(opts *Options) { opts.LogError = log }) }
}

// Sets error handlers to return specific to each error HTTP Status Codes.
func ErrorHandlers[O Option](handlers ...ErrorHandler) func(O) {
	return func(o O) {
		o.Set(func(opts *Options) { opts.ErrorHandlers = append(opts.ErrorHandlers, handlers...) })
	}
}

// Sets response writer.
func ResponseWriter[O Option](w WriteResponse) func(O) {
	return func(o O) { o.Set(func(opts *Options) { opts.WriteResponse = w }) }
}

// Sets requests content reader.
func RequestContentReader(r ReadRequest) func(*ActionOptions) {
	return func(opts *ActionOptions) { opts.ReadRequestContent = r }
}
