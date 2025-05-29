package http

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sony/gobreaker"
)

// nonRepeatableErrorStatuses содержит статус-коды, которые НЕ должны считаться повторяемыми ошибками
// и не должны приводить к открытию circuit breaker
var nonRepeatableErrorStatuses = map[int]struct{}{
	http.StatusInternalServerError: {}, // 500
	http.StatusBadGateway:          {}, // 502
	http.StatusServiceUnavailable:  {}, // 503
	http.StatusRequestTimeout:      {}, // 408
	http.StatusTooManyRequests:     {}, // 429
}

// WithCircuitBreakerMiddleware создает middleware для реализации Circuit Breaker паттерна.
// Принимает конфигурацию для CircuitBreaker и возвращает MiddlewareFunc.
// Middleware будет:
//   - Создавать новый CircuitBreaker с переданными настройками
//   - Отслеживать ошибки запросов
//   - Открывать цепь при превышении лимита ошибок
//   - Возвращать ошибку при открытом контуре
//
// Пример использования:
//   settings := gobreaker.Settings{
//       Name:        "http-client",
//       MaxRequests: 5,
//       Interval:    30 * time.Second,
//       Timeout:     10 * time.Second,
//   }
//   client := NewClient("http://example.com", WithCircuitBreakerMiddleware(settings))
func WithCircuitBreakerMiddleware(settings gobreaker.Settings) MiddlewareFunc {
	cb := gobreaker.NewCircuitBreaker(settings)
	return func(req *http.Request, next func(*http.Request) (*http.Response, error)) (*http.Response, error) {
		res, err := cb.Execute(func() (interface{}, error) {
			resp, err := next(req)
			if err != nil {
				return nil, err
			}


			if _, ok := nonRepeatableErrorStatuses[resp.StatusCode]; ok {
				return nil, &CircuitBreakerError{
					StatusCode: resp.StatusCode,
					Message:    http.StatusText(resp.StatusCode),
				}
			}
			return resp, nil
		})

		if err != nil {
			return nil, err
		}
		return res.(*http.Response), nil
	}
}

// WithRetryMiddleware создает middleware для повторения неудачных запросов
func WithRetryMiddleware(maxRetries int, minWait, maxWait time.Duration) MiddlewareFunc {
	client := retryablehttp.NewClient()
	client.RetryMax = maxRetries
	client.RetryWaitMin = minWait
	client.RetryWaitMax = maxWait
	client.CheckRetry = retryablehttp.DefaultRetryPolicy
	client.Backoff = retryablehttp.DefaultBackoff

	return func(req *http.Request, next func(*http.Request) (*http.Response, error)) (*http.Response, error) {
		// Конвертируем стандартный http.Request в retryablehttp.Request
		retryReq, err := retryablehttp.FromRequest(req)
		if err != nil {
			return nil, err
		}

		// Выполняем запрос с повторами
		resp, err := client.Do(retryReq)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}