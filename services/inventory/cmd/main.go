package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gwall-e/services/inventory/internal/api/cron"
	httpapi "github.com/gwall-e/services/inventory/internal/api/http"
	"github.com/gwall-e/services/inventory/internal/application"
	"github.com/gwall-e/services/inventory/internal/application/commands"
	"github.com/gwall-e/services/inventory/internal/application/queries"
	"github.com/gwall-e/services/inventory/internal/infra/bot"
	"github.com/gwall-e/services/inventory/internal/infra/events"
	invmongo "github.com/gwall-e/services/inventory/internal/infra/mongo"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// =========================================================================
	// Инфраструктура: MongoDB
	// =========================================================================

	mongoCtx, mongoCancel := context.WithTimeout(ctx, 10*time.Second)
	defer mongoCancel()

	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	mongoClient, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("connect to mongodb: %v", err)
	}
	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(shutCtx); err != nil {
			log.Printf("disconnect mongodb: %v", err)
		}
	}()

	db := mongoClient.Database(getEnv("MONGO_DB", "inventory"))

	// =========================================================================
	// Инфраструктура: внешние клиенты
	// =========================================================================

	httpClient := &http.Client{Timeout: 5 * time.Second}
	botInventoryURL := getEnv("BOT_INVENTORY_URL", "http://localhost:9090")
	botAPIKey := getEnv("BOT_INVENTORY_API_KEY", "")

	// =========================================================================
	// Выходные адаптеры: репозитории (infra/mongo)
	// =========================================================================

	projectRepo := invmongo.NewMongoProjectRepository(db)
	hostRepo := invmongo.NewMongoHostRepository(db)
	hostReadModel := invmongo.NewMongoHostReadModel(db)

	// =========================================================================
	// Выходные адаптеры: порты (infra/bot, infra/events)
	// =========================================================================

	botClient := bot.NewHTTPBotInventoryClient(botInventoryURL, botAPIKey, httpClient)
	publisher := events.NewLogEventPublisher()

	// =========================================================================
	// Application: command handlers
	// =========================================================================

	registerHostHandler := commands.NewRegisterHostHandler(hostRepo, projectRepo, publisher)
	syncShadowHostsHandler := commands.NewSyncShadowHostsHandler(nil /* shadowHostRepo */, botClient, publisher)
	provisionHostFromShadowHandler := commands.NewProvisionHostFromShadowHandler(hostRepo, nil /* shadowHostRepo */, projectRepo, publisher)
	createProjectHandler := commands.NewCreateProjectHandler(projectRepo)
	createNamespaceHandler := commands.NewCreateNamespaceHandler(nil /* nsRepo */)
	provisionVMHandler := commands.NewProvisionVMHandler(nil /* vmRepo */, projectRepo)

	// =========================================================================
	// Application: query handlers
	// =========================================================================

	listHostsHandler := queries.NewListHostsHandler(hostReadModel)
	getHostHandler := queries.NewGetHostHandler(hostReadModel)
	getProjectHandler := queries.NewGetProjectHandler(nil /* projectReadModel */)
	listProjectsHandler := queries.NewListProjectsHandler(nil /* projectReadModel */)
	listNamespacesHandler := queries.NewListNamespacesHandler(nil /* nsReadModel */)
	getNamespaceHandler := queries.NewGetNamespaceHandler(nil /* nsReadModel */)

	// =========================================================================
	// Application: CommandDispatcher — write-side (команды в транзакции)
	//
	// Pipeline: TxManager.RunInTx → LoggingMiddleware → TracingMiddleware → Handler
	// txManager == nil: команды выполняются без транзакции (реализация в infra/mongo)
	// =========================================================================

	cmdDispatcher := application.NewCommandDispatcher(
		nil, // txManager — TODO: реализовать в infra/mongo
		application.WithCommandMiddleware(
			application.LoggingMiddleware(),
			application.TracingMiddleware(),
		),
	)

	application.RegisterCommand(cmdDispatcher, registerHostHandler)
	application.RegisterCommand(cmdDispatcher, syncShadowHostsHandler)
	application.RegisterCommand(cmdDispatcher, provisionHostFromShadowHandler)
	application.RegisterCommand(cmdDispatcher, createProjectHandler)
	application.RegisterCommand(cmdDispatcher, createNamespaceHandler)
	application.RegisterCommand(cmdDispatcher, provisionVMHandler)

	// =========================================================================
	// Application: QueryDispatcher — read-side (запросы без транзакции)
	//
	// Pipeline: LoggingMiddleware → TracingMiddleware → Handler
	// =========================================================================

	qryDispatcher := application.NewQueryDispatcher(
		application.WithQueryMiddleware(
			application.LoggingMiddleware(),
			application.TracingMiddleware(),
		),
	)

	application.RegisterQuery(qryDispatcher, listHostsHandler)
	application.RegisterQuery(qryDispatcher, getHostHandler)
	application.RegisterQuery(qryDispatcher, getProjectHandler)
	application.RegisterQuery(qryDispatcher, listProjectsHandler)
	application.RegisterQuery(qryDispatcher, listNamespacesHandler)
	application.RegisterQuery(qryDispatcher, getNamespaceHandler)

	// =========================================================================
	// Входные адаптеры: HTTP (internal/api/http/)
	// Команды → CommandDispatcher, запросы → QueryDispatcher.
	// =========================================================================

	hostsHTTP := httpapi.NewHostsHandler(cmdDispatcher, qryDispatcher)
	projectsHTTP := httpapi.NewProjectsHandler(cmdDispatcher, qryDispatcher)
	namespacesHTTP := httpapi.NewNamespacesHandler(cmdDispatcher, qryDispatcher)

	httpAddr := getEnv("HTTP_ADDR", ":8080")
	httpServer := httpapi.NewServer(httpAddr, hostsHTTP, projectsHTTP, namespacesHTTP)

	// =========================================================================
	// Входные адаптеры: CRON (internal/api/cron/)
	// CRON напрямую использует конкретный хендлер — не нужен Mediator.
	// =========================================================================

	syncInterval := parseDuration(getEnv("SYNC_SHADOW_INTERVAL", "5m"))
	syncJob := cron.NewSyncShadowHostsJob(syncShadowHostsHandler, syncInterval)

	// =========================================================================
	// Запуск
	// =========================================================================

	go syncJob.Start(ctx)

	go func() {
		if err := httpServer.Run(); err != nil {
			log.Printf("http server error: %v", err)
		}
	}()

	log.Printf("inventory service started (http=%s, sync_interval=%s)", httpAddr, syncInterval)

	<-ctx.Done()
	log.Printf("shutting down...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutCtx); err != nil {
		log.Printf("http server shutdown error: %v", err)
	}

	log.Printf("inventory service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("invalid duration %q, using 5m: %v", s, err)
		return 5 * time.Minute
	}
	return d
}
