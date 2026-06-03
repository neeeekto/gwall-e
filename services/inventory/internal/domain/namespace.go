package domain

import (
	"fmt"
	"time"
)

// NamespaceID — типизированный идентификатор namespace.
type NamespaceID string

// Namespace — агрегат инвентаря.
// Namespace — логическое пространство имён внутри проекта.
type Namespace struct {
	id        NamespaceID
	projectID ProjectID
	name      string
	createdAt time.Time
	updatedAt time.Time
}

// NewNamespace — фабричный метод.
func NewNamespace(
	id NamespaceID,
	projectID ProjectID,
	name string,
) (*Namespace, error) {
	if name == "" {
		return nil, ErrInvalidNamespaceName
	}

	now := time.Now().UTC()
	return &Namespace{
		id:        id,
		projectID: projectID,
		name:      name,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// RestoreNamespaceParams — параметры восстановления из хранилища.
type RestoreNamespaceParams struct {
	ID        NamespaceID
	ProjectID ProjectID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RestoreNamespace — конструктор для восстановления из персистентного хранилища.
func RestoreNamespace(p RestoreNamespaceParams) *Namespace {
	return &Namespace{
		id:        p.ID,
		projectID: p.ProjectID,
		name:      p.Name,
		createdAt: p.CreatedAt,
		updatedAt: p.UpdatedAt,
	}
}

// --- Геттеры ---

func (n *Namespace) ID() NamespaceID      { return n.id }
func (n *Namespace) ProjectID() ProjectID { return n.projectID }
func (n *Namespace) Name() string         { return n.name }
func (n *Namespace) CreatedAt() time.Time { return n.createdAt }
func (n *Namespace) UpdatedAt() time.Time { return n.updatedAt }

// --- Доменные методы ---

func (n *Namespace) Rename(name string) error {
	if name == "" {
		return ErrInvalidNamespaceName
	}
	if n.name == name {
		return nil
	}
	n.name = name
	n.updatedAt = time.Now().UTC()
	return nil
}

func (n *Namespace) BelongsToProject(projectID ProjectID) bool {
	return n.projectID == projectID
}

func (n *Namespace) String() string {
	return fmt.Sprintf("Namespace{id=%s, project=%s, name=%s}", n.id, n.projectID, n.name)
}
