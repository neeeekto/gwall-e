package http

import "fmt"

// NonRepeatableError представляет ошибку для статус-кодов, которые не должны повторяться
type NonRepeatableError struct {
	StatusCode int
	Message    string
}

func (e *NonRepeatableError) Error() string {
	return fmt.Sprintf("circuit breaker error: %d %s", e.StatusCode, e.Message)
}