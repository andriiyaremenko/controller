package controller

import (
	"errors"
	"net/http"
)

type ErrorHandler func(error, ReadParam) (int, any)

func getErrorResponse(err error, paramReader ReadParam, handlers []ErrorHandler) (int, any) {
	for _, handle := range handlers {
		code, response := handle(err, paramReader)
		if response != nil {
			return code, response
		}
	}

	return http.StatusInternalServerError, err.Error()
}

func HandleError[E any](target E, httpCode int) ErrorHandler {
	return func(err error, readParam ReadParam) (int, any) {
		if errors.As(err, &target) {
			return httpCode, target
		}

		return 0, nil
	}
}

func HandleErrorAs[E any](target E, httpCode int, as func(E, ReadParam) any) ErrorHandler {
	return func(err error, readParam ReadParam) (int, any) {
		if errors.As(err, &target) {
			return httpCode, as(target, readParam)
		}

		return 0, nil
	}
}
