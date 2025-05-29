package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreakerMiddleware создает middleware для реализации Circuit Breaker паттерна.
// Принимает конфигурацию для CircuitBreaker и возвращает MiddlewareFunc.
// Middleware будет:
//   - Создавать новый CircuitBreaker с переданными настройками
//   - Отслеживать ошибки запросов
//   - Открывать цепь при превышении лимита ошибок
//   - Возвращать ошибку при открытом контуре
//
// Пример использования:
//
//	settings := gobreaker.Settings{
//	    Name:        "http-client",
//	    MaxRequests: 5,
//	    Interval:    30 * time.Second,
//	    Timeout:     10 * time.Second,
//	}
//	client := NewClient("http://example.com",
//	  WithMiddleware(CircuitBreakerMiddleware(settings)),
//	  WithTransport(NewRetryableTransport(3, 1*time.Second, 5*time.Second)),
//	)
type CircuitBreakerConfig struct {
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
	MaxFailures uint32
}

func CircuitBreakerMiddleware(config CircuitBreakerConfig) MiddlewareFunc {
	settings := gobreaker.Settings{
		Name:        "http-client",
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= config.MaxFailures
		},
		IsSuccessful: func(err error) bool {
			var cbErr *CircuitBreakerError
			if errors.As(err, &cbErr) {
				return false
			}
			return true
		},
	}
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
