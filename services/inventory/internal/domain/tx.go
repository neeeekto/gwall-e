package domain

import "context"

// TxManager — порт для управления транзакциями.
// Реализация живёт в infra/mongo (или infra/postgres и т.д.).
//
// Использование в CommandDispatcher:
//
//	err := txManager.RunInTx(ctx, func(ctx context.Context) error {
//	    // ctx содержит активную транзакцию
//	    // все репозитории должны извлекать сессию из ctx
//	    return handler.Handle(ctx, cmd)
//	})
type TxManager interface {
	// RunInTx выполняет fn внутри транзакции.
	// Если fn возвращает ошибку — транзакция откатывается.
	// Если fn завершается успешно — транзакция коммитится.
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
