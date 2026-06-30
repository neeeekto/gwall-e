// Package example содержит throwaway-артефакты для smoke-проверки кодогена mockery (SVC-06).
// Реальных доменных портов в проекте ещё нет — они появятся в Phase 6/7; этот пакет
// существует только чтобы доказать работоспособность генерации моков (mockery v3).
package example

import (
	"context"
	"errors"
)

// ExampleID — типизированный идентификатор примера-ресурса (style.md: типизированные ID
// вместо «голой» строки). Не доменный тип; служит лишь для smoke-демонстрации.
type ExampleID string

// ErrExampleProvisionFailed — sentinel-ошибка примера (style.md: предсказуемые ошибки —
// sentinel, сравниваются через errors.Is). Не доменная ошибка; только для smoke.
// Реализации порта оборачивают её через %w, сохраняя цепочку.
var ErrExampleProvisionFailed = errors.New("example provision failed")

// ExampleProvisioner — throwaway пример-порт для smoke-проверки mockery (SVC-06).
// ЭТО НЕ доменный порт: реальные порты (репозитории, UnitOfWork и т.п.) появятся
// в Phase 6/7. Интерфейс существует, чтобы mockery сгенерировал по нему мок и unit-spec
// доказал рабочий expecter-API.
type ExampleProvisioner interface {
	// Provision выделяет пример-ресурс с заданным именем под идентификатором id;
	// возвращает обёрнутую ErrExampleProvisionFailed при неуспехе.
	Provision(ctx context.Context, id ExampleID, name string) error
}
