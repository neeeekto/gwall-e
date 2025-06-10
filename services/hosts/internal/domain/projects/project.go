package projects

import (
	"context"

	"github.com/gwall-e/hosts/events"
	"github.com/gwall-e/hosts/internal/domain/projects/contracts"
	"github.com/gwall-e/hosts/internal/domain/projects/entities"
	"github.com/gwall-e/hosts/internal/domain/projects/validators"
	"github.com/gwall-e/pkg/core_entities"
	set "github.com/deckarep/golang-set/v2"
	"time"
	"strings"
	"slices"
	"github.com/gwall-e/hosts/internal/domain/projects/errors"
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
	events       []events.Event         `bson:"_"`
}

func (p *Project) Events() []events.Event {
	return p.events
}

func (p *Project) addEvent(event events.Event) {
	event.SetOccurredOn(time.Now().Unix())
	p.events = append(p.events, event)

}

func NewProject(ctx context.Context, checker contracts.ProjectChecker, id string, name string, projectType core_entities.UnitType) (*Project, error) {
	err := validators.ValidateId(ctx, checker, id)
	if err != nil {
		return nil, err
	}

	err = validators.ValidateName(name)
	if err != nil {
		return nil, err
	}

	project := &Project{
		ID:           id,
		Name:         name,
		Type:         projectType,
		Tags:         []string{},
		Description:  "",
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

func (p *Project) SetName(name string) error {
	err := validators.ValidateName(name)
	if err != nil {
		return err
	}

	p.Name = strings.TrimSpace(name)
	p.addEvent(&events.ProjectInfoChangedEvent{ID: p.ID, Name: &name})
	return nil
}

func (p *Project) SetDesc(decs string) error {
	p.Description = strings.TrimSpace(decs)
	p.addEvent(&events.ProjectInfoChangedEvent{ID: p.ID, Description: &decs})
	return nil
}

func (p *Project) SetTags(tags []string) error {
	tagSet := set.NewSet[string]()
	for _, tag := range tags {
		prepared := strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if len(prepared) == 0 {
			continue
		}
		tagSet.Add(prepared)
	}
	res := tagSet.ToSlice()
	if len(res) == 0 && len(tags) > 0 {
		return &errors.ProjectValidationError{
			Field:   "tags",
			Message: "empty tags",
		}
	}

	p.Tags = res
	slices.Sort(p.Tags)
	p.addEvent(&events.ProjectInfoChangedEvent{ID: p.ID, Tags: &p.Tags})
	return nil
}

func (p *Project) SetOwners(owners []string) {
	p.Owners = owners
	p.addEvent(&events.ProjectInfoChangedEvent{ID: p.ID, Owners: &p.Owners})
}
