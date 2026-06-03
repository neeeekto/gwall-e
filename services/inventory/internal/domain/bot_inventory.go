package domain

import "context"

// BotHostRecord — запись о хосте из внешней bot-инвентори.
// Это DTO внешней системы — изолирован от доменной модели через порт.
type BotHostRecord struct {
	ExternalID string
	Inv        int
	Kind       string // "server" | "mac"
	IsMounted  bool
	Location   BotLocation
	Hardware   BotHardware
}

type BotLocation struct {
	Country  string
	City     string
	Building string
	Module   string
	Rack     string
	Unit     string
}

type BotHardware struct {
	Name        string
	Platform    string
	IPMIMac     string
	Motherboard string
	MACs        []string
}

// BotInventoryClient — порт для получения данных из внешней bot-инвентори.
// Определён в домене, реализован в infra/bot.
type BotInventoryClient interface {
	FetchAll(ctx context.Context) ([]BotHostRecord, error)
	FetchByExternalID(ctx context.Context, externalID string) (*BotHostRecord, error)
	FetchUpdatedSince(ctx context.Context, since int64) ([]BotHostRecord, error)
}
