package httpapi

// ============================================================
// Hosts DTOs
// ============================================================

// RegisterHostRequest — тело запроса POST /api/v1/hosts.
type RegisterHostRequest struct {
	ProjectID string              `json:"project_id"`
	FQDN      string              `json:"fqdn"`
	Inv       int                 `json:"inv"`
	Kind      string              `json:"kind"` // "server" | "mac"
	Location  LocationRequest     `json:"location"`
	Hardware  HostHardwareRequest `json:"hardware"`
}

type LocationRequest struct {
	Country  string `json:"country"`
	City     string `json:"city"`
	Building string `json:"building"`
	Module   string `json:"module"`
	Rack     string `json:"rack"`
	Unit     string `json:"unit"`
	Object   string `json:"object"`
	RoomType string `json:"room_type"`
}

type HostHardwareRequest struct {
	Name        string   `json:"name"`
	Platform    string   `json:"platform"`
	IPMIMac     string   `json:"ipmi_mac"`
	Motherboard string   `json:"motherboard"`
	MACs        []string `json:"macs"`
}

// RegisterHostResponse — ответ на POST /api/v1/hosts.
type RegisterHostResponse struct {
	HostID string `json:"host_id"`
}

// ============================================================
// Projects DTOs
// ============================================================

// CreateProjectRequest — тело запроса POST /api/v1/projects.
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Kind        string `json:"kind"` // "baremetal" | "vm"
}

// CreateProjectResponse — ответ на POST /api/v1/projects.
type CreateProjectResponse struct {
	ProjectID string `json:"project_id"`
}

// ============================================================
// Namespaces DTOs
// ============================================================

// CreateNamespaceRequest — тело запроса POST /api/v1/projects/{projectID}/namespaces.
type CreateNamespaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateNamespaceResponse — ответ на POST /api/v1/projects/{projectID}/namespaces.
type CreateNamespaceResponse struct {
	NamespaceID string `json:"namespace_id"`
}

// ============================================================
// Error response
// ============================================================

// ErrorResponse — стандартный формат ошибки API.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}
