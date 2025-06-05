package projects

import (
	"context"

	"github.com/gwall-e/hosts/events"
	"github.com/gwall-e/hosts/internal/domain/projects/contracts"
	"github.com/gwall-e/hosts/internal/domain/projects/entities"
	"github.com/gwall-e/hosts/internal/domain/projects/validators"
	"github.com/gwall-e/pkg/core_entities"
)

type Project struct {
	ID           string                 `bson:"_id"`
	Name         string                 `bson:"name"`
	Type         core_entities.UnitType `bson:"type"`
	Tags         []string               `bson:"tags"`
	Description  string                 `bson:"description"`
	CMS          []entities.CMS         `bson:"cms"`
	Network      *entities.Network      `bson:"network"`
	Deploying    *entities.Deploying    `bson:"deploying"`
	Profiling    *entities.Profiling    `bson:"profiling"`
	Notification *entities.Notification `bson:"notification"`
	Monitoring   *entities.Monitoring   `bson:"monitoring"`
	Task         *entities.Task         `bson:"task"`
	Tier         byte                   `bson:"tier"`
	Owners       []string               `bson:"owners"`
	Inventory    *entities.Inventory    `bson:"inventory"`
	events       []interface{}          `bson:"_"`
}

func (p *Project) Events() []interface{} {
	return p.events
}

func (p *Project) addEvent(event interface{}) {
	p.events = append(p.events, event)
}

func NewProject(ctx context.Context, checker contracts.ProjectChecker, id string, name string, projectType core_entities.UnitType, desc string) (*Project, error) {
	err := validators.ValidateId(ctx, checker, id)
	if err != nil {
		return nil, err
	}

	project := &Project{
		ID:           id,
		Name:         name,
		Type:         projectType,
		Tags:         []string{},
		Description:  desc,
		CMS:          []entities.CMS{},
		Network:      nil,
		Deploying:    nil,
		Profiling:    nil,
		Notification: nil,
		Monitoring:   nil,
		Task:         nil,
		Tier:         0,
		Owners:       []string{},
		Inventory:    nil,
	}

	project.addEvent(&events.ProjectAddedEvent{ID: id, Name: name, Type: projectType})

	return project, nil
}

func (p *Project) SetTags(tags []string) {

}
