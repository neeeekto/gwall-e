package http

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

func NewRetryableTransport(maxRetries int, minWait, maxWait time.Duration) *http.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = maxRetries
	client.RetryWaitMin = minWait
	client.RetryWaitMax = maxWait
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp != nil {
			if _, ok := nonRepeatableErrorStatuses[resp.StatusCode]; ok {
				return false, nil
			}
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}
	client.Backoff = retryablehttp.DefaultBackoff

	return client.StandardClient()
}
