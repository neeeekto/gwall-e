package entities

import (
	"github.com/google/uuid"
	"github.com/gwall-e/hosts/internal/domain/common"
)

type Project struct {
	ID           string          `bson:"_id"`
	Name         string          `bson:"name"`
	Type         common.UnitType `bson:"type"`
	Tags         []string        `bson:"tags"`
	Description  string          `bson:"description"`
	CMS          []CMS           `bson:"cms"`
	Network      *Network        `bson:"network"`
	Deploying    *Deploying      `bson:"deploying"`
	Profiling    *Profiling      `bson:"profiling"`
	Notification *Notification   `bson:"notification"`
	Monitoring   *Monitoring     `bson:"monitoring"`
	Task         *Task           `bson:"task"`
	Tier         byte            `bson:"tier"`
	Owners       []string        `bson:"owners"`
	Inventory    *Inventory      `bson:"inventory"`
}

// NewProject creates a new Project instance
func NewProject(id uuid.UUID, name string, projectType common.UnitType) *Project {
	return &Project{
		ID:        id.String(),
		Name:      name,
		Type:      projectType,
		Tags:      make([]string, 0),
		Owners:    make([]string, 0),
		Inventory: &Inventory{},
	}
}
