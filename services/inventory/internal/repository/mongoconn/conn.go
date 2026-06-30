// Package mongoconn — тонкий RS-aware фабричный helper подключения к MongoDB (D-14).
//
// Эталон «как мы подключаемся к Mongo» для Phase 6/7: client-фабрика на mongo-driver/v2
// с writeconcern.Majority + health-ping. ГРАНИЦА (D-14): здесь НЕТ repository/UnitOfWork/
// Outbox/транзакций — это слой repositories Phase 7 (см. knowledge/architecture.md;
// MUST NOT возрождать TxManager/tx.go).
package mongoconn

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

// Connect создаёт клиент MongoDB по URI и проверяет связь health-ping'ом.
//
// В mongo-driver/v2 mongo.Connect принимает только options (БЕЗ ctx — отличие от v1);
// ctx используется в операциях Ping/Disconnect. Запись настроена на writeconcern.Majority
// (согласованность — core value gwall-e).
//
// Для локального single-node replica set URI ДОЛЖЕН содержать
// "?replicaSet=rs0&directConnection=true" — directConnection обходит RS-discovery, который
// иначе уводит клиент на advertise'нутый из rs.initiate хост (Pitfall 1). Сам helper хост
// НЕ хардкодит: для testcontainers доверяем строке из ConnectionString.
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	cl, err := mongo.Connect(
		options.Client().
			ApplyURI(uri).
			SetWriteConcern(writeconcern.Majority()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to Mongo: %w", err)
	}
	// health-ping: readpref nil = primary; при ошибке закрываем клиент, чтобы не течь соединениями.
	if err := cl.Ping(ctx, nil); err != nil {
		_ = cl.Disconnect(ctx)
		return nil, fmt.Errorf("ping Mongo: %w", err)
	}
	return cl, nil
}
