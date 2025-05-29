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

// NewClient создает и возвращает новый HTTP клиент с настройками.
//
// Параметры:
//   - baseURL: базовый URL для всех запросов (обязательный)
//   - setup: вариативный список функций настройки (опционально):
//     - WithMiddleware: добавляет middleware обработчики
//     - WithTransport: устанавливает кастомный HTTP транспорт
//
// Примеры использования:
//
// 1. Простой клиент без дополнительных настроек:
//    client := NewClient("https://api.example.com")
//
// 2. Клиент с Circuit Breaker:
//    client := NewClient("https://api.example.com",
//      WithMiddleware(CircuitBreakerMiddleware(gobreaker.Settings{
//        Name:        "api-client",
//        MaxRequests: 5,
//        Interval:    30 * time.Second,
//        Timeout:     10 * time.Second,
//      })),
//    )
//
// 3. Клиент с Retry и Circuit Breaker:
//    client := NewClient("https://api.example.com",
//      WithMiddleware(CircuitBreakerMiddleware(gobreaker.Settings{})),
//      WithTransport(NewRetryableTransport(3, 1*time.Second, 5*time.Second)),
//    )
//
// 4. Клиент с кастомным транспортом:
//    client := NewClient("https://api.example.com",
//      WithTransport(&http.Client{
//        Timeout: 30 * time.Second,
//      }),
//    )
//
// 5. Комбинированный клиент:
//    client := NewClient("https://api.example.com",
//      WithMiddleware(
//        CircuitBreakerMiddleware(gobreaker.Settings{}),
//        YourCustomMiddleware(),
//      ),
//      WithTransport(NewRetryableTransport(3, 1*time.Second, 5*time.Second)),
//    )
//
// Возвращает:
//   - Реализацию интерфейса HTTPClient, готовую к использованию
//   - При ошибках в middleware или транспорте может вернуть ошибку
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