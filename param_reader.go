package controller

import "net/http"

const (
	FromHeaders ParamSource = iota // Reads from Headers.
	FromURL                        // Reads from URL parameters.
)

// Callback to read from request parameters.
type ReadParam func(ParamSource, string) string

// Request parameter source.
type ParamSource int

func readParam(
	req *http.Request,
	readURLParam func(*http.Request, string) string,
) func(ParamSource, string) string {
	return func(source ParamSource, key string) string {
		switch source {
		case FromHeaders:
			return req.Header.Get(key)
		case FromURL:
			return readURLParam(req, key)
		default:
			return ""
		}
	}
}
