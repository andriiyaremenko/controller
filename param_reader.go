package controller

import "net/http"

// Reads from Headers.
func FromHeaders(req *http.Request, key string) string {
	return req.Header.Get(key)
}
