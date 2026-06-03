package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// botHostResponse — DTO ответа от bot-инвентори API.
type botHostResponse struct {
	ID        string      `json:"id"`
	Inv       int         `json:"inv"`
	Kind      string      `json:"kind"`
	IsMounted bool        `json:"is_mounted"`
	Location  botLocation `json:"location"`
	Hardware  botHardware `json:"hardware"`
}

type botLocation struct {
	Country  string `json:"country"`
	City     string `json:"city"`
	Building string `json:"building"`
	Module   string `json:"module"`
	Rack     string `json:"rack"`
	Unit     string `json:"unit"`
}

type botHardware struct {
	Name        string   `json:"name"`
	Platform    string   `json:"platform"`
	IPMIMac     string   `json:"ipmi_mac"`
	Motherboard string   `json:"motherboard"`
	MACs        []string `json:"macs"`
}

// HTTPBotInventoryClient — HTTP-клиент к bot-инвентори.
// Реализует domain.BotInventoryClient.
type HTTPBotInventoryClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewHTTPBotInventoryClient(baseURL, apiKey string, httpClient *http.Client) *HTTPBotInventoryClient {
	return &HTTPBotInventoryClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		apiKey:     apiKey,
	}
}

var _ domain.BotInventoryClient = (*HTTPBotInventoryClient)(nil)

func (c *HTTPBotInventoryClient) FetchAll(ctx context.Context) ([]domain.BotHostRecord, error) {
	return c.fetch(ctx, fmt.Sprintf("%s/api/v1/hosts", c.baseURL))
}

func (c *HTTPBotInventoryClient) FetchByExternalID(ctx context.Context, externalID string) (*domain.BotHostRecord, error) {
	url := fmt.Sprintf("%s/api/v1/hosts/%s", c.baseURL, externalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call bot inventory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bot inventory returned status %d", resp.StatusCode)
	}

	var dto botHostResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	record := translateRecord(dto)
	return &record, nil
}

func (c *HTTPBotInventoryClient) FetchUpdatedSince(ctx context.Context, since int64) ([]domain.BotHostRecord, error) {
	url := fmt.Sprintf("%s/api/v1/hosts?updated_since=%d", c.baseURL, since)
	return c.fetch(ctx, url)
}

func (c *HTTPBotInventoryClient) fetch(ctx context.Context, url string) ([]domain.BotHostRecord, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call bot inventory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bot inventory returned status %d", resp.StatusCode)
	}

	var dtos []botHostResponse
	if err := json.NewDecoder(resp.Body).Decode(&dtos); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	records := make([]domain.BotHostRecord, 0, len(dtos))
	for _, dto := range dtos {
		records = append(records, translateRecord(dto))
	}
	return records, nil
}

// translateRecord — ядро ACL: переводит внешнюю модель bot во внутреннюю доменную.
func translateRecord(dto botHostResponse) domain.BotHostRecord {
	return domain.BotHostRecord{
		ExternalID: dto.ID,
		Inv:        dto.Inv,
		Kind:       dto.Kind,
		IsMounted:  dto.IsMounted,
		Location: domain.BotLocation{
			Country:  dto.Location.Country,
			City:     dto.Location.City,
			Building: dto.Location.Building,
			Module:   dto.Location.Module,
			Rack:     dto.Location.Rack,
			Unit:     dto.Location.Unit,
		},
		Hardware: domain.BotHardware{
			Name:        dto.Hardware.Name,
			Platform:    dto.Hardware.Platform,
			IPMIMac:     dto.Hardware.IPMIMac,
			Motherboard: dto.Hardware.Motherboard,
			MACs:        dto.Hardware.MACs,
		},
	}
}
