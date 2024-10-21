/*
This package is aimed to reduce amount of boilerplate code when writing regular http.Handle.
It provides simple and explicit API to define HTTP endpoints.

To install controller:

	go get -u github.com/andriiyaremenko/controller

How to use:
package main

import (

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/andriiyaremenko/controller"

)

// type definitions and other code..

	func main() {
		r := chi.NewRouter()
		handle := func(r *http.Request) (ResponseModel, error) {
			service := SomeService(r.Context())
			result, err := service.Do()

			if err != nil {
				return result, err
			}

			return result, nil
		}

		r.Post(
			"/", controller.
				Respond[ResponseModel](handle).
				With(controller.SuccessCode(http.StatusCreated)),
		)

		http.ListenAndServe(":3000", r)
	}

```
*/
package controller
