package domain

import "errors"

// Ошибки хостов
var (
	ErrHostAlreadyExists     = errors.New("host already exists")
	ErrHostNotFound          = errors.New("host not found")
	ErrInvalidFQDN           = errors.New("fqdn cannot be empty")
	ErrInvalidInv            = errors.New("inv number must be positive")
	ErrInvalidProjectForHost = errors.New("project does not support bare-metal hosts")
	ErrInvalidHostKind       = errors.New("invalid host kind")
	ErrInvalidStatus         = errors.New("invalid status transition")
)

// Ошибки проектов
var (
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrProjectNotFound      = errors.New("project not found")
	ErrInvalidProjectName   = errors.New("project name cannot be empty")
	ErrInvalidProjectKind   = errors.New("invalid project kind")
)

// Ошибки namespace
var (
	ErrNamespaceAlreadyExists = errors.New("namespace already exists")
	ErrNamespaceNotFound      = errors.New("namespace not found")
	ErrInvalidNamespaceName   = errors.New("namespace name cannot be empty")
)

// Ошибки VM
var (
	ErrVMAlreadyExists     = errors.New("virtual machine already exists")
	ErrVMNotFound          = errors.New("virtual machine not found")
	ErrInvalidVMName       = errors.New("vm name cannot be empty")
	ErrInvalidVMSpec       = errors.New("invalid vm specification")
	ErrInvalidProjectForVM = errors.New("project does not support virtual machines")
	ErrInvalidVMStatus     = errors.New("invalid vm status transition")
)

// Ошибки физической иерархии
var (
	ErrDataCenterNotFound      = errors.New("datacenter not found")
	ErrDataCenterAlreadyExists = errors.New("datacenter already exists")
	ErrInvalidDataCenterName   = errors.New("datacenter name cannot be empty")
	ErrModuleNotFound          = errors.New("module not found")
	ErrModuleAlreadyExists     = errors.New("module already exists")
	ErrInvalidModuleName       = errors.New("module name cannot be empty")
	ErrRackNotFound            = errors.New("rack not found")
	ErrRackAlreadyExists       = errors.New("rack already exists")
	ErrInvalidRackName         = errors.New("rack name cannot be empty")
	ErrInvalidRackCapacity     = errors.New("rack capacity must be positive")
)
