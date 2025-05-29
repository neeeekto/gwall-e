package http

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// NewRetryableTransport создает http.Client с поддержкой повторных запросов
func NewRetryableTransport(maxRetries int, minWait, maxWait time.Duration) *http.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = maxRetries
	client.RetryWaitMin = minWait
	client.RetryWaitMax = maxWait
	client.CheckRetry = retryablehttp.DefaultRetryPolicy
	client.Backoff = retryablehttp.DefaultBackoff
	
	return client.StandardClient()
}