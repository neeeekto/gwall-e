package domain

import (
	"fmt"
	"time"
)

// ProjectStatus — жизненный цикл проекта.
type ProjectStatus int

const (
	ProjectStatusActive ProjectStatus = iota
	ProjectStatusArchived
)

func (s ProjectStatus) String() string {
	switch s {
	case ProjectStatusActive:
		return "active"
	case ProjectStatusArchived:
		return "archived"
	default:
		return "unknown"
	}
}

// Project — агрегат инвентаря.
// Ключевой инвариант: тип проекта (Kind) задаётся при создании и не меняется.
type Project struct {
	id          ProjectID
	name        string
	description string
	kind        ProjectKind
	tags        []string
	status      ProjectStatus
	createdAt   time.Time
	updatedAt   time.Time
}

// NewProject — фабричный метод.
func NewProject(
	id ProjectID,
	name string,
	description string,
	kind ProjectKind,
) (*Project, error) {
	if name == "" {
		return nil, ErrInvalidProjectName
	}
	if kind == 0 {
		return nil, ErrInvalidProjectKind
	}

	now := time.Now().UTC()
	return &Project{
		id:          id,
		name:        name,
		description: description,
		kind:        kind,
		status:      ProjectStatusActive,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// RestoreProjectParams — параметры восстановления из хранилища.
type RestoreProjectParams struct {
	ID          ProjectID
	Name        string
	Description string
	Kind        ProjectKind
	Tags        []string
	Status      ProjectStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RestoreProject — конструктор для восстановления из персистентного хранилища.
func RestoreProject(p RestoreProjectParams) *Project {
	return &Project{
		id:          p.ID,
		name:        p.Name,
		description: p.Description,
		kind:        p.Kind,
		tags:        p.Tags,
		status:      p.Status,
		createdAt:   p.CreatedAt,
		updatedAt:   p.UpdatedAt,
	}
}

// --- Геттеры ---

func (p *Project) ID() ProjectID         { return p.id }
func (p *Project) Name() string          { return p.name }
func (p *Project) Description() string   { return p.description }
func (p *Project) Kind() ProjectKind     { return p.kind }
func (p *Project) Status() ProjectStatus { return p.status }
func (p *Project) CreatedAt() time.Time  { return p.createdAt }
func (p *Project) UpdatedAt() time.Time  { return p.updatedAt }

func (p *Project) Tags() []string {
	if p.tags == nil {
		return nil
	}
	result := make([]string, len(p.tags))
	copy(result, p.tags)
	return result
}

// --- Доменные методы ---

func (p *Project) IsBareMetal() bool {
	return p.kind == ProjectKindBareMetal
}

func (p *Project) IsVM() bool {
	return p.kind == ProjectKindVM
}

func (p *Project) Archive() error {
	if p.status == ProjectStatusArchived {
		return fmt.Errorf("project is already archived")
	}
	p.status = ProjectStatusArchived
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Project) Rename(name string) error {
	if name == "" {
		return ErrInvalidProjectName
	}
	p.name = name
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Project) AddTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}
	for _, t := range p.tags {
		if t == tag {
			return nil
		}
	}
	p.tags = append(p.tags, tag)
	p.updatedAt = time.Now().UTC()
	return nil
}
