package http

import (
	"context"
	"io"
	"net/http"
)

type MiddlewareFunc func(*http.Request, func(*http.Request) (*http.Response, error)) (*http.Response, error)

type httpClient struct {
	baseURL     string
	client      *http.Client
	middleware  []MiddlewareFunc
}


func NewClient(baseURL string, middleware ...MiddlewareFunc) HTTPClient {
	return &httpClient{
		baseURL:    baseURL,
		client:     &http.Client{},
		middleware: middleware,
	}
}

func (c *httpClient) applyMiddleware(req *http.Request) (*http.Response, error) {
	handler := func(r *http.Request) (*http.Response, error) {
		return c.client.Do(r)
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