package validators

import (
	"context"
	"
)

func ValidateId(ctx context.Context, checker contracts.ProjectChecker, id string) error {
	if len(id) > MAX_ID_LENGT {
		return &errors.ProjectValidationError{
			Field:   "id",
			Message: "id is too long",
		}
	}
	if id == "" {
		return &errors.ProjectValidationError{
			Field:   "id",
			Message: "id is required",
		}
	}
	exists, err := checker.CheckIdUnique(ctx, id)
	if err != nil {
		return err
	}

	if exists {
		return &errors.ProjectValidationError{
			Field:   "id",
			Message: "project with this id already exists",
		}
	}
	return nil
}
