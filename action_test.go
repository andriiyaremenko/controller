package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/andriiyaremenko/controller"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type testError struct {
	Detail string `json:"detail"`
}

func (e *testError) Error() string {
	return e.Detail
}

var _ = Describe("Action", func() {
	requestBody := `"Hello World"`
	h := func(
		_ context.Context, greet string, _ func(controller.ParamSource, string) string,
	) (string, error) {
		Expect(greet).To(Equal("Hello World"))

		return "success", nil
	}

	It("should work with defaults", func() {
		action := controller.Action[string, string](h)
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("success"))
	})

	It("should use WithHTTPStatusCode option", func() {
		action := controller.
			Action[string, string](h).
			With(controller.HTTPStatus[*controller.ActionOptions](http.StatusCreated))
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("success"))
	})

	It("should use WithURLParamReader option", func() {
		h := func(
			_ context.Context, greet string, readParam func(controller.ParamSource, string) string,
		) (string, error) {
			Expect(greet).To(Equal("Hello World"))
			Expect(readParam(controller.FromURL, "test")).To(Equal("test"))

			return "success", nil
		}
		action := controller.
			Action[string, string](h).
			With(
				controller.URLParamReader[*controller.ActionOptions](
					func(_ *http.Request, key string) string {
						if key == "test" {
							return "test"
						}

						return ""
					},
				),
			)
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("success"))
	})

	It("should be able to read from Headers", func() {
		h := func(
			_ context.Context, greet string, readParam func(controller.ParamSource, string) string,
		) (string, error) {
			Expect(greet).To(Equal("Hello World"))
			Expect(readParam(controller.FromHeaders, "Test")).To(Equal("test"))

			return "success", nil
		}
		action := controller.Action[string, string](h)
		ts := httptest.NewServer(action)

		defer ts.Close()

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s", ts.URL), strings.NewReader(requestBody))

		req.Header.Add("Content-Type", "application/json; charset=utf-8")
		req.Header.Add("Test", "test")

		resp, err := http.DefaultClient.Do(req)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("success"))
	})

	It("should use WithAppError option", func() {
		h := func(
			_ context.Context, greet string, _ func(controller.ParamSource, string) string,
		) (string, error) {
			Expect(greet).To(Equal("Hello World"))

			return "", &testError{Detail: "oops"}
		}
		action := controller.
			Action[string, string](h).
			With(
				controller.ErrorHandlers[*controller.ActionOptions](
					controller.IfError[*testError](http.StatusBadRequest),
				),
			)
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result testError

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result.Detail).To(Equal("oops"))
	})

	It("should use WithAppError option with mapping", func() {
		h := func(
			_ context.Context, greet string, _ func(controller.ParamSource, string) string,
		) (string, error) {
			Expect(greet).To(Equal("Hello World"))

			return "", fmt.Errorf("oooh")
		}
		action := controller.
			Action[string, string](h).
			With(
				controller.ErrorHandlers[*controller.ActionOptions](
					controller.IfError[*testError](http.StatusBadRequest),
					controller.IfErrorUse(
						func(err error, _ controller.ReadParam) any {
							return &testError{Detail: err.Error()}
						},
						http.StatusConflict,
					),
				),
			)
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusConflict))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result testError

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result.Detail).To(Equal("oooh"))
	})

	It("should use WithErrorLogger option", func() {
		h := func(
			_ context.Context, greet string, _ func(controller.ParamSource, string) string,
		) (string, error) {
			Expect(greet).To(Equal("Hello World"))

			return "", &testError{Detail: "oops"}
		}
		action := controller.
			Action[string, string](h).
			With(
				controller.ErrorHandlers[*controller.ActionOptions](
					controller.IfError[*testError](http.StatusBadRequest),
				),
				controller.ErrorLogger[*controller.ActionOptions](
					func(_ context.Context, err error, message string) {
						Expect(err).Should(BeAssignableToTypeOf(new(testError)))
						Expect(err.(*testError).Detail).To(Equal("oops"))
						Expect(message).To(Equal("request failed"))
					}),
			)
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result testError

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result.Detail).To(Equal("oops"))
	})

	It("should use WithRequestContentReader option", func() {
		called := false
		action := controller.Action[string, string](h).
			With(
				controller.RequestContentReader(func(req *http.Request, model any) error {
					called = true
					return controller.JSONBodyReader(req, model)
				}),
			)
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(called).To(BeTrue())

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("success"))
	})

	It("should use WithResponseWriter option", func() {
		called := false
		action := controller.
			Action[string, string](h).
			With(
				controller.ResponseWriter[*controller.ActionOptions](func(
					ctx context.Context, w http.ResponseWriter,
					logError func(context.Context, error, string),
					status int, data any,
				) {
					called = true
					controller.JSONWriter(ctx, w, logError, status, data)
				}),
			)
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(called).To(BeTrue())

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("success"))
	})

	It("should use Defaults option", func() {
		requestReaderCalled := false
		responseWriterCalled := false
		defaults := controller.Defaults(
			controller.RequestContentReader(func(req *http.Request, model any) error {
				requestReaderCalled = true
				return controller.JSONBodyReader(req, model)
			}),
			controller.ResponseWriter[*controller.ActionOptions](func(
				ctx context.Context, w http.ResponseWriter,
				logError func(context.Context, error, string),
				status int, data any,
			) {
				responseWriterCalled = true
				controller.JSONWriter(ctx, w, logError, status, data)
			}),
		)

		action := controller.
			Action[string, string](h).
			With(
				defaults,
				controller.HTTPStatus[*controller.ActionOptions](http.StatusCreated),
			)

		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(requestReaderCalled).To(BeTrue())
		Expect(responseWriterCalled).To(BeTrue())

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("success"))
	})
})
