package http_test

import (
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sony/gobreaker"

	. "github.com/gwall-e/pkg/http"
)

var _ = Describe("CircuitBreakerMiddleware", func() {
	var (
		middleware  MiddlewareFunc
		nextHandler func(*http.Request) (*http.Response, error)
		req         *http.Request
	)

	BeforeEach(func() {
		config := CircuitBreakerConfig{
			MaxRequests: 1,
			Interval:    1 * time.Second,
			Timeout:     1 * time.Second,
			MaxFailures: 1,
		}
		middleware = CircuitBreakerMiddleware(config)
		req, _ = http.NewRequest("GET", "http://example.com", nil)
	})

	Context("when next handler returns non-repeatable error status", func() {
		BeforeEach(func() {
			nextHandler = func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
				}, nil
			}
		})

		It("should return CircuitBreakerError", func() {
			resp, err := middleware(req, nextHandler)
			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&CircuitBreakerError{}))
		})
	})

	Context("when next handler returns regular error", func() {
		BeforeEach(func() {
			nextHandler = func(r *http.Request) (*http.Response, error) {
				return nil, errors.New("connection error")
			}
		})

		It("should return the error", func() {
			resp, err := middleware(req, nextHandler)
			Expect(resp).To(BeNil())
			Expect(err).To(MatchError("connection error"))
		})
	})

	Context("when next handler succeeds", func() {
		BeforeEach(func() {
			nextHandler = func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
				}, nil
			}
		})

		It("should return the response", func() {
			resp, err := middleware(req, nextHandler)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("when exceeding max failures", func() {
		var callCount int

		BeforeEach(func() {
			callCount = 0
			nextHandler = func(r *http.Request) (*http.Response, error) {
				callCount++
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
				}, nil
			}
		})

		It("should open circuit after max failures", func() {
			// First call - failure
			_, err := middleware(req, nextHandler)
			Expect(err).To(BeAssignableToTypeOf(&CircuitBreakerError{}))

			// Second call - circuit should be open
			_, err = middleware(req, nextHandler)
			Expect(err).To(MatchError(gobreaker.ErrOpenState))

			Expect(callCount).To(Equal(1)) // Only first call should reach handler
		})
	})
})
