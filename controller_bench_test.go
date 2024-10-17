package controller_test

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/andriiyaremenko/controller"
)

type testResponseWriter struct{}

// Header implements http.ResponseWriter.
func (t testResponseWriter) Header() http.Header {
	return make(http.Header)
}

// Write implements http.ResponseWriter.
func (t testResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

// WriteHeader implements http.ResponseWriter.
func (t testResponseWriter) WriteHeader(statusCode int) {}

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func BenchmarkControllerHandleSuccess(b *testing.B) {
	requestBody := `"Hello World"`
	h := func(r *http.Request) (*string, error) {
		return controller.ReadJSON[string](r)
	}
	_, err := http.NewRequest("POST", "/test", strings.NewReader(requestBody))
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	action := controller.
		Handle[*string](h).
		With(
			controller.SuccessCode(http.StatusCreated),
			controller.HandleErrorWithCode[*testError](http.StatusBadRequest),
		)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("POST", "/test", strings.NewReader(requestBody))
		action.ServeHTTP(testResponseWriter{}, r)
	}
}

func BenchmarkControllerHandleError(b *testing.B) {
	requestBody := `"Hello World"`
	h := func(_ *http.Request) (*string, error) {
		return nil, &testError{Detail: "test"}
	}
	_, err := http.NewRequest("POST", "/test", strings.NewReader(requestBody))
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	action := controller.
		Handle[*string](h).
		With(
			controller.SuccessCode(http.StatusCreated),
			controller.HandleErrorWithCode[*testError](http.StatusBadRequest),
		)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("POST", "/test", strings.NewReader(requestBody))
		action.ServeHTTP(testResponseWriter{}, r)
	}
}

func Benchmark_100_000_ParallelControllerHandleSuccess(b *testing.B) {
	requestBody := `"Hello World"`
	h := func(r *http.Request) (*string, error) {
		return controller.ReadJSON[string](r)
	}
	_, err := http.NewRequest("POST", "/test", strings.NewReader(requestBody))
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	action := controller.
		Handle[*string](h).
		With(
			controller.SuccessCode(http.StatusCreated),
			controller.HandleErrorWithCode[*testError](http.StatusBadRequest),
		)

	b.ResetTimer()
	b.SetParallelism(100_000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, _ := http.NewRequest("POST", "/test", strings.NewReader(requestBody))
			action.ServeHTTP(testResponseWriter{}, r)
		}
	})
}

func Benchmark_100_000_ParallelControllerHandleError(b *testing.B) {
	requestBody := `"Hello World"`
	h := func(_ *http.Request) (*string, error) {
		return nil, &testError{Detail: "test"}
	}
	_, err := http.NewRequest("POST", "/test", strings.NewReader(requestBody))
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	action := controller.
		Handle[*string](h).
		With(
			controller.SuccessCode(http.StatusCreated),
			controller.HandleErrorWithCode[*testError](http.StatusBadRequest),
		)

	b.ResetTimer()
	b.SetParallelism(100_000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, _ := http.NewRequest("POST", "/test", strings.NewReader(requestBody))
			action.ServeHTTP(testResponseWriter{}, r)
		}
	})
}
