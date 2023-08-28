package controller

import (
	"context"
	"net/http"
)

// Shared Action and Task options.
type Options struct {
	ReadRequestParam map[string]func(*http.Request, string) string
	LogError         func(context.Context, error, string)
	WriteResponse    func(
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

// Task options.
type TaskOptions struct {
	Options
}

// Action options.
type ActionOptions struct {
	Options

	ReadRequestContent func(*http.Request, any) error
}

// Task and Action Options union type.
type Option interface {
	*Options | *TaskOptions | *ActionOptions
	Set(func(*Options))
}

// Sets success response HTTP Status Code.
func HTTPStatus[O Option](code int) func(O) {
	return func(o O) { o.Set(func(opts *Options) { opts.SuccessCode = code }) }
}

// Sets URL parameters reader.
func RequestParam[O Option](key string, fn func(*http.Request, string) string) func(O) {
	return func(o O) { o.Set(func(opts *Options) { opts.ReadRequestParam[key] = fn }) }
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

// Option to transform generic func(*controller.Options)
// to func(*controller.ActionOptions) or func(*controller.TaskOptions)
func As[O Option](opts func(*Options)) func(O) {
	return func(o O) { o.Set(opts) }
}

// Option to bind several options into one to share as default settings.
func Defaults[O Option](opts ...func(O)) func(O) {
	return func(o O) {
		for _, opt := range opts {
			opt(o)
		}
	}
}

// Sets requests content reader.
func RequestContentReader(r ReadRequest) func(*ActionOptions) {
	return func(opts *ActionOptions) { opts.ReadRequestContent = r }
}
