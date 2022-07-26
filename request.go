package controller

import (
	"encoding/json"
	"io"
	"net/http"
)

type RequestReader func(*http.Request, any) error

func JSONBodyReader(req *http.Request, model any) error {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return &ReadRequestError{err: err}
	}

	if err := json.Unmarshal(b, &model); err != nil {
		return &ReadRequestError{err: err}
	}

	return nil
}
