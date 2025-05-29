package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTP Client", func() {
	var (
		client    HTTPClient
		testServer *httptest.Server
	)

	BeforeEach(func() {
		client = NewClient("")
	})

	Context("HTTP Methods", func() {
		BeforeEach(func() {
			testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case "GET":
					w.WriteHeader(http.StatusOK)
				case "POST":
					body, _ := io.ReadAll(r.Body)
					if string(body) == "test body" {
						w.WriteHeader(http.StatusCreated)
					}
				case "PUT":
					body, _ := io.ReadAll(r.Body)
					if string(body) == "test body" {
						w.WriteHeader(http.StatusOK)
					}
				case "DELETE":
					w.WriteHeader(http.StatusNoContent)
				case "PATCH":
					body, _ := io.ReadAll(r.Body)
					if string(body) == "test body" {
						w.WriteHeader(http.StatusOK)
					}
				}
			}))
			client = NewClient(testServer.URL)
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("should perform GET requests", func() {
			resp, err := client.Get(context.Background(), "/", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should perform POST requests with body", func() {
			body := strings.NewReader("test body")
			resp, err := client.Post(context.Background(), "/", body, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		})

		It("should perform PUT requests with body", func() {
			body := strings.NewReader("test body")
			resp, err := client.Put(context.Background(), "/", body, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should perform DELETE requests", func() {
			resp, err := client.Delete(context.Background(), "/", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("should perform PATCH requests with body", func() {
			body := strings.NewReader("test body")
			resp, err := client.Patch(context.Background(), "/", body, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("Middleware", func() {
		BeforeEach(func() {
			testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Test") == "test" {
					w.WriteHeader(http.StatusOK)
				}
			}))

			middleware := func(req *http.Request, next func(*http.Request) (*http.Response, error)) (*http.Response, error) {
				req.Header.Set("X-Test", "test")
				return next(req)
			}

			client = NewClient(testServer.URL, WithMiddleware(middleware))
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("should apply middleware", func() {
			resp, err := client.Get(context.Background(), "/", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("Custom Requests", func() {
		BeforeEach(func() {
			testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "CUSTOM" {
					w.WriteHeader(http.StatusOK)
				}
			}))
			client = NewClient(testServer.URL)
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("should handle custom requests with Do", func() {
			req, err := http.NewRequest("CUSTOM", testServer.URL, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := client.Do(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})
	
	Context("Transport Configuration", func() {
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
				WithMiddleware(middleware),
				WithTransport(customClient),
			)

			_, err := client.Get(context.Background(), "/", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(middlewareCalled).To(BeTrue())
			Expect(client.(*httpClient).transport).To(Equal(customClient))
		})
	})

	Context("Transport Configuration", func() {
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
				WithMiddleware(middleware),
				WithTransport(customClient),
			)

			_, err := client.Get(context.Background(), "/", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(middlewareCalled).To(BeTrue())
			Expect(client.(*httpClient).transport).To(Equal(customClient))
		})
	})
})