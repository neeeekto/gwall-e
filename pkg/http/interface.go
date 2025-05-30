package http

import (
	"context"
	"io"
	"net/http"
)

type HTTPClient interface {
	Get(ctx context.Context, path string, queryParams map[string]string, headers map[string]string) (*http.Response, error)
	Post(ctx context.Context, path string, body io.Reader, headers map[string]string) (*http.Response, error)
	Put(ctx context.Context, path string, body io.Reader, headers map[string]string) (*http.Response, error)
	Delete(ctx context.Context, path string, headers map[string]string) (*http.Response, error)
	Patch(ctx context.Context, path string, body io.Reader, headers map[string]string) (*http.Response, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}
