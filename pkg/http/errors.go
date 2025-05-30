package http

import (
	"errors"
	"fmt"
)

var (
	// ErrCircuitBreakOpen возвращается когда circuit breaker открыт
	ErrCircuitBreakOpen = errors.New("circuit breaker is open")
)

// NonRepeatableError представляет ошибку для статус-кодов, которые не должны повторяться
type NonRepeatableError struct {
	StatusCode int
	Message    string
}

func (e *NonRepeatableError) Error() string {
	return fmt.Sprintf("circuit breaker error: %d %s", e.StatusCode, e.Message)
}
