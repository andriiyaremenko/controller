# controller

This package is aimed to reduce amount of boilerplate code when writing regular http.Handle.
It provides simple and explicit API to define HTTP endpoints.

### To install controller:
`go get -u github.com/andriiyaremenko/controller`

### How to use:
```go
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/andriiyaremenko/controller"
)

// type definitions and other code..

func main() {
	r := chi.NewRouter()
	action := func(
		ctx context.Context, requestModel MyModel, readParam func(controller.ParamSource, string) string,
	) (ResponseModel, error) {
		service := SomeService(ctx)
		result, err := service.Do(requestModel, readParam(command.FromHeader, "request-id"))

		if err != nil {
			return result, err
		}

		return result, nil
	}
	task := func(
		ctx context.Context, readParam func(controller.ParamSource, string) string,
	) (ResponseModel, error) {
		service := SomeQuery(ctx)
		result, err := service.Do(readParam(command.FromHeader, "request-id"))

		if err != nil {
			return result, err
		}

		return result, nil
	}
	logError := func(_ context.Context, err error, message string) {
		log.Printf("error: %s, message: %s\n", err, message)
	}

	r.Post(
		"/", controller.
			Action[MyModel, ResponseModel](action).
			With(
				controller.URLParamReader[*controller.ActionOptions](chi.URLParam),
				controller.ErrorHandlers[*controller.ActionOptions](
					controller.IfError[*testError](http.StatusBadRequest),
					controller.IfErrorUse(
						func(err error, _ controller.ReadParam) any {
							return &testError{Detail: err.Error()}
						},
						http.StatusConflict,
					),
				),
				controller.ErrorLogger[*controller.ActionOptions](logError),
				controller.HTTPStatus[*controller.ActionOptions](http.StatusCreated),
		),
	)
	r.Get(
		"/", controller.
			Task[ResponseModel](task).
			With(
				controller.URLParamReader[*controller.TaskOptions](chi.URLParam),
				controller.ErrorHandlers[*controller.TaskOptions](
					controller.IfError[*testError](http.StatusBadRequest),
					controller.IfErrorUse(
						func(err error, _ controller.ReadParam) any {
							return &testError{Detail: err.Error()}
						},
						http.StatusConflict,
					),
				),
				controller.ErrorLogger[*controller.TaskOptions](logError),
				controller.HTTPStatus[*controller.TaskOptions](http.StatusCreated),
		),
	)

	http.ListenAndServe(":3000", r)
```
