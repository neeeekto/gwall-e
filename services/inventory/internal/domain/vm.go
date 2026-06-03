package domain

import (
	"fmt"
	"time"
)

// VMID — типизированный идентификатор виртуальной машины.
type VMID string

// VMStatus — жизненный цикл виртуальной машины.
type VMStatus int

const (
	VMStatusProvisioning VMStatus = iota
	VMStatusRunning
	VMStatusStopped
	VMStatusTerminated
)

func (s VMStatus) String() string {
	switch s {
	case VMStatusProvisioning:
		return "provisioning"
	case VMStatusRunning:
		return "running"
	case VMStatusStopped:
		return "stopped"
	case VMStatusTerminated:
		return "terminated"
	default:
		return "unknown"
	}
}

// VM — агрегат инвентаря.
// Ключевой инвариант: VM может существовать только в VM-проекте.
type VM struct {
	id          VMID
	projectID   ProjectID
	namespaceID NamespaceID
	name        string
	spec        VMSpec
	tags        []string
	status      VMStatus
	createdAt   time.Time
	updatedAt   time.Time
}

// NewVM — фабричный метод.
func NewVM(
	id VMID,
	projectID ProjectID,
	namespaceID NamespaceID,
	name string,
	spec VMSpec,
) (*VM, error) {
	if name == "" {
		return nil, ErrInvalidVMName
	}
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &VM{
		id:          id,
		projectID:   projectID,
		namespaceID: namespaceID,
		name:        name,
		spec:        spec,
		status:      VMStatusProvisioning,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// RestoreVMParams — параметры восстановления из хранилища.
type RestoreVMParams struct {
	ID          VMID
	ProjectID   ProjectID
	NamespaceID NamespaceID
	Name        string
	Spec        VMSpec
	Tags        []string
	Status      VMStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RestoreVM — конструктор для восстановления из персистентного хранилища.
func RestoreVM(p RestoreVMParams) *VM {
	return &VM{
		id:          p.ID,
		projectID:   p.ProjectID,
		namespaceID: p.NamespaceID,
		name:        p.Name,
		spec:        p.Spec,
		tags:        p.Tags,
		status:      p.Status,
		createdAt:   p.CreatedAt,
		updatedAt:   p.UpdatedAt,
	}
}

// --- Геттеры ---

func (v *VM) ID() VMID                 { return v.id }
func (v *VM) ProjectID() ProjectID     { return v.projectID }
func (v *VM) NamespaceID() NamespaceID { return v.namespaceID }
func (v *VM) Name() string             { return v.name }
func (v *VM) Spec() VMSpec             { return v.spec }
func (v *VM) Status() VMStatus         { return v.status }
func (v *VM) CreatedAt() time.Time     { return v.createdAt }
func (v *VM) UpdatedAt() time.Time     { return v.updatedAt }

func (v *VM) Tags() []string {
	if v.tags == nil {
		return nil
	}
	result := make([]string, len(v.tags))
	copy(result, v.tags)
	return result
}

// --- Доменные методы ---

func (v *VM) MarkRunning() error {
	if v.status != VMStatusProvisioning {
		return fmt.Errorf("%w: cannot mark running vm with status %s", ErrInvalidVMStatus, v.status)
	}
	v.status = VMStatusRunning
	v.updatedAt = time.Now().UTC()
	return nil
}

func (v *VM) Stop() error {
	if v.status != VMStatusRunning {
		return fmt.Errorf("%w: cannot stop vm with status %s", ErrInvalidVMStatus, v.status)
	}
	v.status = VMStatusStopped
	v.updatedAt = time.Now().UTC()
	return nil
}

func (v *VM) Start() error {
	if v.status != VMStatusStopped {
		return fmt.Errorf("%w: cannot start vm with status %s", ErrInvalidVMStatus, v.status)
	}
	v.status = VMStatusRunning
	v.updatedAt = time.Now().UTC()
	return nil
}

func (v *VM) Terminate() error {
	if v.status == VMStatusTerminated {
		return fmt.Errorf("%w: vm is already terminated", ErrInvalidVMStatus)
	}
	v.status = VMStatusTerminated
	v.updatedAt = time.Now().UTC()
	return nil
}

func (v *VM) Resize(spec VMSpec) error {
	if v.status != VMStatusStopped {
		return fmt.Errorf("%w: vm must be stopped before resize", ErrInvalidVMStatus)
	}
	if err := spec.Validate(); err != nil {
		return err
	}
	v.spec = spec
	v.updatedAt = time.Now().UTC()
	return nil
}

func (v *VM) AddTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}
	for _, t := range v.tags {
		if t == tag {
			return nil
		}
	}
	v.tags = append(v.tags, tag)
	v.updatedAt = time.Now().UTC()
	return nil
}
