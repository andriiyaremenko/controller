# controller

This package is aimed to reduce amount of boilerplate code when writing regular http.Handle.
It provides simple and explicit API to define HTTP endpoints.

### To install controller:
`go get -u github.com/andriiyaremenko/controller`

### How to use:
```go
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/andriiyaremenko/controller"
)

// type definitions and other code..

func main() {
	r := chi.NewRouter()
	action := func(ctx context.Context, requestModel MyModel) (ResponseModel, error) {
		service := SomeService(ctx)
		result, err := service.Do(requestModel, controller.ContextParam(ctx, "request-id"))

		if err != nil {
			return result, err
		}

		return result, nil
	}
	task := func(ctx context.Context) (ResponseModel, error) {
		service := SomeQuery(ctx)
		result, err := service.Do(controller.ContextParam(ctx, "request-id"))

		if err != nil {
			return result, err
		}

		return result, nil
	}
	logError := func(_ context.Context, err error, message string) {
		log.Printf("error: %s, message: %s\n", err, message)
	}
	defaults := controller.Defaults(
		controller.RequestParam[*controller.Options]("request-id", chi.URLParam),
		controller.ErrorHandlers[*controller.Options](
			controller.IfError[*testError](http.StatusBadRequest),
			controller.IfErrorUse(
				func(_ context.Context, err error) any {
					return &testError{Detail: err.Error()}
				},
				http.StatusConflict,
			),
		),
		controller.ErrorLogger[*controller.Options](logError),
	)

	r.Post(
		"/", controller.
			Action[MyModel, ResponseModel](action).
			With(
				controller.As[*controller.ActionOptions](defaults),
				controller.HTTPStatus[*controller.ActionOptions](http.StatusCreated),
			),
	)
	r.Get(
		"/", controller.
			Task[ResponseModel](task).
			With(controller.As[*controller.TaskOptions](defaults)),
	)

	http.ListenAndServe(":3000", r)
}
```
