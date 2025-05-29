package http

import (
	"context"
	"io"
	"net/http"
)

// MiddlewareFunc определяет тип функции middleware
type MiddlewareFunc func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)

// SetupFunc определяет тип функции для настройки клиента
type SetupFunc func(*httpClient)

type httpClient struct {
	baseURL     string
	transport   *http.Client
	middleware  []MiddlewareFunc
}

// WithMiddleware добавляет middleware в клиент
func WithMiddleware(middleware ...MiddlewareFunc) SetupFunc {
	return func(c *httpClient) {
		c.middleware = append(c.middleware, middleware...)
	}
}

// WithTransport устанавливает кастомный http.Client
func WithTransport(transport *http.Client) SetupFunc {
	return func(c *httpClient) {
		c.transport = transport
	}
}

// NewClient создает новый HTTP клиент с настройками.
// Принимает:
//   - baseURL: базовый URL для всех запросов
//   - setup: опциональные функции настройки (WithMiddleware, WithTransport)
//
// Примеры использования:
//   // Простой клиент без middleware
//   client := NewClient("http://example.com")
//
//   // Клиент с Circuit Breaker
//   client := NewClient("http://example.com",
//     WithMiddleware(WithCircuitBreakerMiddleware(gobreaker.Settings{})),
//   )
//
//   // Клиент с Retry и Circuit Breaker
//   client := NewClient("http://example.com",
//     WithMiddleware(WithCircuitBreakerMiddleware(gobreaker.Settings{})),
//     WithTransport(NewRetryableTransport(3, 1*time.Second, 5*time.Second)),
//   )
//
//   // Клиент с кастомным транспортом
//   client := NewClient("http://example.com",
//     WithTransport(&http.Client{Timeout: 30*time.Second}),
//   )
//
// Возвращает реализацию HTTPClient.
func NewClient(baseURL string, setup ...SetupFunc) HTTPClient {
	c := &httpClient{
		baseURL:   baseURL,
		transport: &http.Client{},
	}

	for _, fn := range setup {
		fn(c)
	}

	return c
}

func (c *httpClient) applyMiddleware(req *http.Request) (*http.Response, error) {
	handler := func(r *http.Request) (*http.Response, error) {
		return c.transport.Do(r)
	}

	for i := len(c.middleware) - 1; i >= 0; i-- {
		mw := c.middleware[i]
		next := handler
		handler = func(r *http.Request) (*http.Response, error) {
			return mw(r, next)
		}
	}

	return handler(req)
}

func (c *httpClient) Get(ctx context.Context, path string, queryParams map[string]string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", normalizeURL(c.baseURL, path), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	
	if queryParams != nil {
		q := req.URL.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.applyMiddleware(req)
}

func (c *httpClient) Post(ctx context.Context, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", normalizeURL(c.baseURL, path), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.applyMiddleware(req)
}

func (c *httpClient) Put(ctx context.Context, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("PUT", normalizeURL(c.baseURL, path), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.applyMiddleware(req)
}

func (c *httpClient) Delete(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", normalizeURL(c.baseURL, path), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.applyMiddleware(req)
}

func (c *httpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return c.applyMiddleware(req)
}

func (c *httpClient) Patch(ctx context.Context, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("PATCH", normalizeURL(c.baseURL, path), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.applyMiddleware(req)
}