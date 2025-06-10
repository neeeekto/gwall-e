package validators

import (
	"context"
	"github.com/gwall-e/hosts/internal/domain/projects/contracts"
	"github.com/gwall-e/hosts/internal/domain/projects/errors"
)

func ValidateId(ctx context.Context, checker contracts.ProjectChecker, id string) error {
	if len(id) > MaxIdLength {
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

func ValidateName(name string) error {
	if name == "" {
		return &errors.ProjectValidationError{
			Field:   "name",
			Message: "name is required",
		}
	}

	if len(name) > MaxNameLength {
		return &errors.ProjectValidationError{
			Field:   "name",
			Message: "name too long",
		}
	}
	return nil
}
