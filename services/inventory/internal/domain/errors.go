// Package domain содержит доменное ядро Inventory: типизированные ID-VO, агрегаты,
// доменные события и порты. Слой не знает о транспорте, БД или брокере (гексагон) —
// здесь только инварианты предметной области.
package domain

import (
	"errors"
	"fmt"
)

// Доменные sentinel-ошибки — предсказуемые исходы инвариантов, сравниваемые через
// errors.Is (style.md §«Sentinel vs обёрнутые»). Реализации портов оборачивают их
// через %w, сохраняя цепочку.
var (
	// ErrInvalidID — некорректный идентификатор (битая/пустая uuid-строка при parse).
	ErrInvalidID = errors.New("invalid id")
	// ErrInvalidTransition — недопустимый переход жизненного цикла агрегата.
	ErrInvalidTransition = errors.New("invalid lifecycle transition")
	// ErrAlreadyDecommissioned — действие над уже выведенным из эксплуатации хостом.
	ErrAlreadyDecommissioned = errors.New("host already decommissioned")
)

// ErrFQDNConflict — типизированный конфликт уникальности FQDN среди active-хостов
// (D-11/Pitfall 7). Это доменная ошибка, а НЕ сырой DB E11000: реализация БД не
// протекает наружу. Candidates — подсказка советочного матчинга (D-12), может быть пуст.
type ErrFQDNConflict struct {
	FQDN       string   // конфликтующий полностью квалифицированный домен
	ExistingID HostID   // ID уже занявшего этот FQDN active-хоста
	Candidates []HostID // кандидаты на ре-идентификацию (advisory, опционально)
}

// Error реализует интерфейс error (строка — английская, style.md).
func (e ErrFQDNConflict) Error() string {
	return fmt.Sprintf("fqdn %q already taken by active host %s", e.FQDN, e.ExistingID)
}

// ErrProjectNotEmpty — попытка удалить проект, в котором остались хосты (INV-10).
// Доменный контракт delete-only-if-empty; не протекает деталями БД.
type ErrProjectNotEmpty struct {
	ProjectID ProjectID // идентификатор непустого проекта
	HostCount int       // сколько хостов мешают удалению
}

// Error реализует интерфейс error (строка — английская, style.md).
func (e ErrProjectNotEmpty) Error() string {
	return fmt.Sprintf("project %s is not empty: %d host(s) still attached", e.ProjectID, e.HostCount)
}
