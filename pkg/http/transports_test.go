package http_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gwall-e/pkg/http"
)

var _ = Describe("RetryableTransport", func() {
	Context("with non-repeatable status codes", func() {
		It("should not retry requests", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer ts.Close()

			client := NewRetryableTransport(3, 1*time.Millisecond, 10*time.Millisecond)
			
			req, err := http.NewRequest("GET", ts.URL, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
		})
	})

	Context("with repeatable status codes", func() {
		It("should retry requests", func() {
			callCount := 0
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if callCount < 3 {
					w.WriteHeader(http.StatusGatewayTimeout)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			client := NewRetryableTransport(3, 1*time.Millisecond, 10*time.Millisecond)
			
			req, err := http.NewRequest("GET", ts.URL, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(callCount).To(Equal(3))
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})
})