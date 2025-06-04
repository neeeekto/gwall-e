package projects

import (
	"context"
	"github.com/gwall-e/hosts/internal/domain/common"
)

type AddProjectDTO struct {
	Name string         
	Type common.UnitType 
}

func (c *ProjectService) AddProject(ctx context.Context, data *AddProjectDTO) error {
	return nil
}
