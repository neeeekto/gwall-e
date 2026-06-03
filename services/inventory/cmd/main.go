package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gwall-e/services/inventory/internal/application/commands"
	"github.com/gwall-e/services/inventory/internal/application/queries"
	"github.com/gwall-e/services/inventory/internal/infra/bot"
	"github.com/gwall-e/services/inventory/internal/infra/events"
	invmongo "github.com/gwall-e/services/inventory/internal/infra/mongo"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// --- Инфраструктура: MongoDB ---
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("connect to mongodb: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("disconnect mongodb: %v", err)
		}
	}()

	db := mongoClient.Database(getEnv("MONGO_DB", "inventory"))

	// --- Инфраструктура: HTTP-клиент для bot-инвентори ---
	httpClient := &http.Client{Timeout: 5 * time.Second}
	botInventoryURL := getEnv("BOT_INVENTORY_URL", "http://localhost:9090")
	botAPIKey := getEnv("BOT_INVENTORY_API_KEY", "")

	// =========================================================================
	// Репозитории (infra/mongo)
	// =========================================================================

	projectRepo := invmongo.NewMongoProjectRepository(db)
	hostRepo := invmongo.NewMongoHostRepository(db)
	hostReadModel := invmongo.NewMongoHostReadModel(db)

	// =========================================================================
	// Порты (infra/bot, infra/events)
	// =========================================================================

	botClient := bot.NewHTTPBotInventoryClient(botInventoryURL, botAPIKey, httpClient)
	publisher := events.NewLogEventPublisher()

	// =========================================================================
	// Command handlers
	// =========================================================================

	// hosts
	registerHostHandler := commands.NewRegisterHostHandler(hostRepo, projectRepo, publisher)
	syncShadowHostsHandler := commands.NewSyncShadowHostsHandler(nil /* shadowHostRepo */, botClient, publisher)
	provisionHostFromShadowHandler := commands.NewProvisionHostFromShadowHandler(hostRepo, nil /* shadowHostRepo */, projectRepo, publisher)

	// projects
	createProjectHandler := commands.NewCreateProjectHandler(projectRepo)

	// namespaces
	createNamespaceHandler := commands.NewCreateNamespaceHandler(nil /* nsRepo */)

	// vms
	provisionVMHandler := commands.NewProvisionVMHandler(nil /* vmRepo */, projectRepo)

	// =========================================================================
	// Query handlers
	// =========================================================================

	listHostsHandler := queries.NewListHostsHandler(hostReadModel)
	getHostHandler := queries.NewGetHostHandler(hostReadModel)

	getProjectHandler := queries.NewGetProjectHandler(nil /* projectReadModel */)
	listProjectsHandler := queries.NewListProjectsHandler(nil /* projectReadModel */)

	listNamespacesHandler := queries.NewListNamespacesHandler(nil /* nsReadModel */)
	getNamespaceHandler := queries.NewGetNamespaceHandler(nil /* nsReadModel */)

	// =========================================================================
	// Логирование wire-up (в production здесь HTTP/gRPC сервер)
	// =========================================================================

	log.Printf("inventory service started")
	log.Printf("commands: register_host=%T sync_shadow=%T provision_from_shadow=%T",
		registerHostHandler, syncShadowHostsHandler, provisionHostFromShadowHandler)
	log.Printf("commands: create_project=%T create_namespace=%T provision_vm=%T",
		createProjectHandler, createNamespaceHandler, provisionVMHandler)
	log.Printf("queries: list_hosts=%T get_host=%T",
		listHostsHandler, getHostHandler)
	log.Printf("queries: get_project=%T list_projects=%T",
		getProjectHandler, listProjectsHandler)
	log.Printf("queries: list_namespaces=%T get_namespace=%T",
		listNamespacesHandler, getNamespaceHandler)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
