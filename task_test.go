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
	h := func(context.Context, func(controller.ParamSource, string) string) (string, error) {
		return "success", nil
	}

	It("should work with defaults", func() {
		task := controller.Task(h)
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
		task := controller.Task(
			h, controller.WithHTTPStatus(http.StatusCreated),
		)
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
		h := func(
			_ context.Context, readParam func(controller.ParamSource, string) string,
		) (string, error) {
			Expect(readParam(controller.FromURL, "test")).To(Equal("test"))

			return "success", nil
		}
		task := controller.Task(
			h, controller.WithURLParamReader(func(_ *http.Request, key string) string {
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
		h := func(
			_ context.Context, readParam func(controller.ParamSource, string) string,
		) (string, error) {
			Expect(readParam(controller.FromHeaders, "Test")).To(Equal("test"))

			return "success", nil
		}
		task := controller.Task(h)
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
		h := func(context.Context, func(controller.ParamSource, string) string) (string, error) {
			return "", &testError{Detail: "oops"}
		}
		task := controller.Task(h, controller.WithAppError[*testError](http.StatusBadRequest))
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

	It("should use WithErrorLogger option", func() {
		h := func(
			context.Context, func(controller.ParamSource, string) string,
		) (string, error) {
			return "", &testError{Detail: "oops"}
		}
		task := controller.Task(
			h,
			controller.WithAppError[*testError](http.StatusBadRequest),
			controller.WithErrorLogger(func(_ context.Context, err error, message string) {
				Expect(err).Should(BeAssignableToTypeOf(new(testError)))
				Expect(err.(*testError).Detail).To(Equal("oops"))
				Expect(message).To(Equal("request failed"))
			}),
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

	It("should use WithRequestContentReader option", func() {
		called := false
		task := controller.Task(
			h,
			controller.WithRequestContentReader(func(req *http.Request, model any) error {
				called = true
				return controller.JSONBodyReader(req, model)
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

	It("should use WithResponseWriter option", func() {
		called := false
		task := controller.Task(
			h,
			controller.WithResponseWriter(func(
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