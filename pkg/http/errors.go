package http

import "fmt"

// CircuitBreakerError представляет ошибку, которая может привести к размыканию circuit breaker
type CircuitBreakerError struct {
	StatusCode int
	Message    string
}

func (e *CircuitBreakerError) Error() string {
	return fmt.Sprintf("circuit breaker error: %d %s", e.StatusCode, e.Message)
}