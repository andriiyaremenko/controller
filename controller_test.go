// nolint: typecheck
package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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

var _ = Describe("Respond", func() {
	controller.SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))

	requestBody := `"Hello World"`
	h := func(r *http.Request) (string, error) {
		greet, err := controller.ReadJSON[string](r)
		Expect(err).NotTo(HaveOccurred())
		Expect(*greet).To(Equal("Hello World"))

		return "success", nil
	}

	It("with defaults", func() {
		action := controller.Respond[string](h)
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

	It("with SuccessCode option", func() {
		action := controller.
			Respond[string](h).
			With(controller.SuccessCode(http.StatusCreated))
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

	It("with ErrorWithCode option", func() {
		h := func(r *http.Request) (string, error) {
			greet, err := controller.ReadJSON[string](r)
			Expect(err).NotTo(HaveOccurred())
			Expect(*greet).To(Equal("Hello World"))

			return "", &testError{Detail: "oops"}
		}
		action := controller.
			Respond[string](h).
			With(controller.ErrorWithCode[*testError](http.StatusBadRequest))
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

	It("with ErrorHandle option", func() {
		h := func(r *http.Request) (string, error) {
			greet, err := controller.ReadJSON[string](r)
			Expect(err).NotTo(HaveOccurred())
			Expect(*greet).To(Equal("Hello World"))

			return "", fmt.Errorf("oooh")
		}
		action := controller.
			Respond[string](h).
			With(
				controller.ErrorHandle(
					controller.MatchError(func(err error) (any, int) {
						if err.Error() == "oooh" {
							return &testError{Detail: err.Error()}, http.StatusConflict
						}

						return nil, 0
					}),
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

	It("with custom ResponseWriter option", func() {
		called := false
		action := controller.
			Respond[string](h).
			With(
				controller.ResponseWriter(func(ctx context.Context, w http.ResponseWriter, data any, status int) {
					called = true
					controller.WriteJSON(ctx, w, data, status)
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

	It("with status only if nothing was returned", func() {
		h := func(r *http.Request) (any, error) {
			return nil, nil
		}
		action := controller.
			Respond[any](h).
			With(controller.SuccessCode(http.StatusCreated))
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
		Expect(b).Should(BeEmpty())
	})

	It("with InternalServerError if error was not handled", func() {
		h := func(r *http.Request) (any, error) {
			return nil, fmt.Errorf("oh no!")
		}
		action := controller.
			Respond[any](h).
			With(controller.SuccessCode(http.StatusCreated))
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/json; charset=utf-8",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("oh no!"))
	})

	It("with recover from panic", func() {
		h := func(r *http.Request) (string, error) {
			panic(&testError{Detail: "oops"})
		}
		action := controller.
			Respond[string](h).
			With(controller.ErrorWithCode[*testError](http.StatusBadRequest))
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

	It("with bad request error if got malformed request", func() {
		requestBody := "field1=value1&field2=value2"
		h := func(r *http.Request) (any, error) {
			_, err := controller.ReadJSON[string](r)
			return nil, err
		}
		action := controller.Respond[any](h)
		ts := httptest.NewServer(action)

		defer ts.Close()

		resp, err := http.Post(
			fmt.Sprintf("%s", ts.URL),
			"application/x-www-form-urlencoded",
			strings.NewReader(requestBody),
		)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)

		Expect(err).ShouldNot(HaveOccurred())

		var result string

		Expect(json.Unmarshal(b, &result)).ShouldNot(HaveOccurred())
		Expect(result).To(Equal("failed to read request: invalid character 'i' in literal false (expecting 'a')"))
	})

	It("with default error handlers option", func() {
		h := func(r *http.Request) (string, error) {
			greet, err := controller.ReadJSON[string](r)
			Expect(err).NotTo(HaveOccurred())
			Expect(*greet).To(Equal("Hello World"))

			return "", fmt.Errorf("oooh")
		}

		controller.SetDefaultErrorHandlers(
			controller.MatchError(func(err error) (any, int) {
				if err.Error() == "oooh" {
					return &testError{Detail: err.Error()}, http.StatusConflict
				}

				return nil, 0
			}),
		)

		action := controller.Respond[string](h)
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
})
