package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/andriiyaremenko/controller"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task", func() {
	h := func(context.Context) (string, error) {
		return "success", nil
	}

	It("should work with defaults", func() {
		task := controller.Task[string](h)
		ts := httptest.NewServer(task)

		defer ts.Close()

		resp, err := http.Get(fmt.Sprintf("%s", ts.URL))

		Expect(err).ShouldNot(HaveOccurred())

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("success"))
	})

	It("should use WithHTTPStatusCode option", func() {
		task := controller.
			Task[string](h).
			With(controller.HTTPStatus(http.StatusCreated))
		ts := httptest.NewServer(task)

		defer ts.Close()

		resp, err := http.Get(fmt.Sprintf("%s", ts.URL))

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
		h := func(ctx context.Context) (string, error) {
			Expect(controller.ContextParam(ctx, "test")).To(Equal("test"))

			return "success", nil
		}
		task := controller.
			Task[string](h).
			With(
				controller.RequestParam(
					"test",
					func(_ *http.Request, key string) string {
						if key == "test" {
							return "test"
						}

						return ""
					}),
			)
		ts := httptest.NewServer(task)

		defer ts.Close()

		resp, err := http.Get(fmt.Sprintf("%s", ts.URL))

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
		h := func(ctx context.Context) (string, error) {
			Expect(controller.ContextParam(ctx, "Test")).To(Equal("test"))

			return "success", nil
		}
		task := controller.
			Task[string](h).
			With(controller.RequestParam("Test", controller.FromHeaders))
		ts := httptest.NewServer(task)

		defer ts.Close()

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s", ts.URL), nil)

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
		h := func(context.Context) (string, error) {
			return "", &testError{Detail: "oops"}
		}
		task := controller.
			Task[string](h).
			With(controller.ErrorHandlers(
				controller.IfError[*testError](http.StatusBadRequest)),
			)
		ts := httptest.NewServer(task)

		defer ts.Close()

		resp, err := http.Get(fmt.Sprintf("%s", ts.URL))

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
		h := func(_ context.Context) (string, error) {
			return "", fmt.Errorf("oooh")
		}
		action := controller.
			Task[string](h).
			With(
				controller.ErrorHandlers(
					controller.IfError[*testError](http.StatusBadRequest),
					controller.IfErrorUse(
						func(_ context.Context, err error) any {
							return &testError{Detail: err.Error()}
						},
						http.StatusConflict,
					),
				),
			)
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Get(fmt.Sprintf("%s", ts.URL))

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
		h := func(context.Context) (string, error) {
			return "", &testError{Detail: "oops"}
		}
		task := controller.
			Task[string](h).
			With(
				controller.ErrorHandlers(
					controller.IfError[*testError](http.StatusBadRequest),
				),
				controller.ErrorLogger(
					func(_ context.Context, err error, message string) {
						Expect(err).Should(BeAssignableToTypeOf(new(testError)))
						Expect(err.(*testError).Detail).To(Equal("oops"))
						Expect(message).To(Equal("request failed"))
					},
				),
			)
		ts := httptest.NewServer(task)

		defer ts.Close()

		resp, err := http.Get(fmt.Sprintf("%s", ts.URL))

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result testError

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result.Detail).To(Equal("oops"))
	})

	It("should use WithResponseWriter option", func() {
		called := false
		task := controller.
			Task[string](h).
			With(
				controller.ResponseWriter(func(
					ctx context.Context, w http.ResponseWriter,
					logError func(context.Context, error, string),
					status int, data any,
				) {
					called = true
					controller.JSONWriter(ctx, w, logError, status, data)
				}),
			)
		ts := httptest.NewServer(task)

		defer ts.Close()

		resp, err := http.Get(fmt.Sprintf("%s", ts.URL))

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
})
