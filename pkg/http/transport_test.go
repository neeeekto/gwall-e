package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/sony/gobreaker"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTP Client Transport", func() {
	var (
		client    HTTPClient
		testServer *httptest.Server
	)

	BeforeEach(func() {
		testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	})

	AfterEach(func() {
		testServer.Close()
	})

	It("should use custom transport", func() {
		customClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		client = NewClient(testServer.URL, WithTransport(customClient))
		Expect(client.(*httpClient).transport).To(Equal(customClient))
	})

	It("should combine middleware and transport", func() {
		var middlewareCalled bool
		middleware := func(req *http.Request, next func(*http.Request) (*http.Response, error)) (*http.Response, error) {
			middlewareCalled = true
			return next(req)
		}

		customClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		client = NewClient(testServer.URL, 
			WithMiddleware(CircuitBreakerMiddleware(gobreaker.Settings{})),
			WithMiddleware(middleware),
			WithTransport(customClient),
		)

		_, err := client.Get(context.Background(), "/", nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(middlewareCalled).To(BeTrue())
		Expect(client.(*httpClient).transport).To(Equal(customClient))
	})
})