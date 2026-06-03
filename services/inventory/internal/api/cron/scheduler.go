package cron

import (
	"context"
	"log"
	"time"

	"github.com/gwall-e/services/inventory/internal/application/commands"
)

// SyncShadowHostsJob — периодическая задача синхронизации shadow-хостов из bot-инвентори.
type SyncShadowHostsJob struct {
	handler  *commands.SyncShadowHostsHandler
	interval time.Duration
}

// NewSyncShadowHostsJob создаёт job с заданным интервалом.
func NewSyncShadowHostsJob(
	handler *commands.SyncShadowHostsHandler,
	interval time.Duration,
) *SyncShadowHostsJob {
	return &SyncShadowHostsJob{
		handler:  handler,
		interval: interval,
	}
}

// Start запускает периодическое выполнение job'а.
// Блокирует до отмены ctx.
func (j *SyncShadowHostsJob) Start(ctx context.Context) {
	log.Printf("cron: sync_shadow_hosts started, interval=%s", j.interval)

	// Первый запуск сразу при старте
	j.run(ctx)

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("cron: sync_shadow_hosts stopped")
			return
		case <-ticker.C:
			j.run(ctx)
		}
	}
}

// run выполняет одну итерацию синхронизации.
func (j *SyncShadowHostsJob) run(ctx context.Context) {
	log.Printf("cron: sync_shadow_hosts running")

	result, err := j.handler.Handle(ctx, commands.SyncShadowHostsCommand{
		Since: 0, // полная синхронизация; для инкрементальной передать unix timestamp
	})
	if err != nil {
		log.Printf("cron: sync_shadow_hosts error: %v", err)
		return
	}

	log.Printf("cron: sync_shadow_hosts done: created=%d updated=%d mounted=%d",
		result.Created, result.Updated, result.Mounted)
}
