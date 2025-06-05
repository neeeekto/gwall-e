package errors

import (
	"fmt"
)

type ProjectValidationError struct {
	Field   string
	Message string
}

func (e ProjectValidationError) Error() string {
	return fmt.Sprintf("project validation error, field: %s, err: %s", e.Field, e.Message)
}
