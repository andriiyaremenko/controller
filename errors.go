package controller

import (
	"net/http"
)

type ErrorHandler func(error) (int, any)

func getErrorResponse(err error, handlers []ErrorHandler) (int, any) {
	for _, handle := range handlers {
		code, response := handle(err)
		if response != nil {
			return code, response
		}
	}

	return http.StatusInternalServerError, err.Error()
}
