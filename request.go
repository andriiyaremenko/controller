package controller

import (
	"encoding/json"
	"io"
	"net/http"
)

// Request reader to read JSON from Body.
func ReadJSON[T any](req *http.Request) (*T, error) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, &ReadRequestError{err: err}
	}

	var model T
	if err := json.Unmarshal(b, &model); err != nil {
		return nil, &ReadRequestError{err: err}
	}

	return &model, nil
}
