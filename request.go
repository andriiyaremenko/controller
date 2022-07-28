package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

// Request reader type.
type RequestReader func(*http.Request, any) error

// Request reader to read JSON from Body.
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

// Request reader to read from Form Data using decode callback.
func FormReader(decode func(any, url.Values) error) RequestReader {
	return func(req *http.Request, model any) error {
		if err := req.ParseForm(); err != nil {
			return &ReadRequestError{err: err}
		}

		if err := decode(model, req.PostForm); err != nil {
			return &ReadRequestError{err: err}
		}

		return nil
	}
}
