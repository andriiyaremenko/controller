package controller

import "net/http"

const (
	FromHeaders ParamSource = iota
	FromURL
)

type ReadParam func(ParamSource, string) string
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
