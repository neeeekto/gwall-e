package domain

import "context"

// HostRepository — интерфейс репозитория хостов.
type HostRepository interface {
	// Save сохраняет новый хост. Возвращает ErrHostAlreadyExists при дубликате.
	Save(ctx context.Context, host *Host) error

	// Update обновляет существующий хост.
	Update(ctx context.Context, host *Host) error

	// FindByID возвращает хост по ID. Возвращает ErrHostNotFound если не найден.
	FindByID(ctx context.Context, id HostID) (*Host, error)

	// ExistsByFQDN проверяет уникальность FQDN в рамках всего инвентаря.
	ExistsByFQDN(ctx context.Context, fqdn string) (bool, error)

	// Delete удаляет хост по ID.
	Delete(ctx context.Context, id HostID) error
}

// ShadowHostRepository — репозиторий для shadow-хостов.
type ShadowHostRepository interface {
	// Save сохраняет новый shadow-хост.
	Save(ctx context.Context, sh *ShadowHost) error

	// Update обновляет существующий shadow-хост.
	Update(ctx context.Context, sh *ShadowHost) error

	// FindByID возвращает shadow-хост по внутреннему ID.
	FindByID(ctx context.Context, id ShadowHostID) (*ShadowHost, error)

	// FindByExternalID возвращает shadow-хост по ID из bot-инвентори.
	FindByExternalID(ctx context.Context, externalID string) (*ShadowHost, error)

	// FindMounted возвращает все смонтированные shadow-хосты.
	FindMounted(ctx context.Context) ([]*ShadowHost, error)
}

// ProjectRepository — интерфейс репозитория проектов.
type ProjectRepository interface {
	// Save сохраняет новый проект. Возвращает ErrProjectAlreadyExists при дубликате.
	Save(ctx context.Context, project *Project) error

	// Update обновляет существующий проект.
	Update(ctx context.Context, project *Project) error

	// FindByID возвращает проект по ID. Возвращает ErrProjectNotFound если не найден.
	FindByID(ctx context.Context, id ProjectID) (*Project, error)

	// ExistsByName проверяет уникальность имени проекта.
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// NamespaceRepository — интерфейс репозитория namespace.
type NamespaceRepository interface {
	// Save сохраняет новый namespace. Возвращает ErrNamespaceAlreadyExists при дубликате.
	Save(ctx context.Context, ns *Namespace) error

	// Update обновляет существующий namespace.
	Update(ctx context.Context, ns *Namespace) error

	// FindByID возвращает namespace по ID. Возвращает ErrNamespaceNotFound если не найден.
	FindByID(ctx context.Context, id NamespaceID) (*Namespace, error)

	// ExistsByName проверяет уникальность имени namespace в рамках проекта.
	ExistsByName(ctx context.Context, projectID ProjectID, name string) (bool, error)

	// Delete удаляет namespace по ID.
	Delete(ctx context.Context, id NamespaceID) error
}

// VMRepository — интерфейс репозитория виртуальных машин.
type VMRepository interface {
	// Save сохраняет новую VM. Возвращает ErrVMAlreadyExists при дубликате.
	Save(ctx context.Context, vm *VM) error

	// Update обновляет существующую VM.
	Update(ctx context.Context, vm *VM) error

	// FindByID возвращает VM по ID. Возвращает ErrVMNotFound если не найдена.
	FindByID(ctx context.Context, id VMID) (*VM, error)

	// ExistsByName проверяет уникальность имени VM в рамках проекта.
	ExistsByName(ctx context.Context, projectID ProjectID, name string) (bool, error)

	// Delete удаляет VM по ID.
	Delete(ctx context.Context, id VMID) error
}

// DataCenterRepository — интерфейс репозитория дата-центров.
type DataCenterRepository interface {
	Save(ctx context.Context, dc *DataCenter) error
	Update(ctx context.Context, dc *DataCenter) error
	FindByID(ctx context.Context, id DataCenterID) (*DataCenter, error)
	Delete(ctx context.Context, id DataCenterID) error
}

// ModuleRepository — интерфейс репозитория модулей ДЦ.
type ModuleRepository interface {
	Save(ctx context.Context, m *Module) error
	Update(ctx context.Context, m *Module) error
	FindByID(ctx context.Context, id ModuleID) (*Module, error)
	FindByDataCenter(ctx context.Context, dcID DataCenterID) ([]*Module, error)
	Delete(ctx context.Context, id ModuleID) error
}

// RackRepository — интерфейс репозитория стоек.
type RackRepository interface {
	Save(ctx context.Context, r *Rack) error
	Update(ctx context.Context, r *Rack) error
	FindByID(ctx context.Context, id RackID) (*Rack, error)
	FindByModule(ctx context.Context, moduleID ModuleID) ([]*Rack, error)
	Delete(ctx context.Context, id RackID) error
}
