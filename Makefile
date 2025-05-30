.PHONY: build run test build-agents build-ui run-ui-dev

build: build-agents build-ui
	@echo "Building all services..."
	@cd services/host_task && go build
	@cd services/auto_healing && go build
	@cd services/host_manager && go build
	@cd services/scenario && go build
	@cd services/audio_logs && go build

build-apps:
	@echo "Building all apps..."
	@cd apps/gwall-e-agent && go build

build-ui:
	@echo "Building UI..."
	@cd gwall-e-ui && npm install && npm run build

run-ui-dev:
	@echo "Running UI in development mode..."
	@cd gwall-e-ui && npx nx serve dashboard

run:
	@echo "Running all services..."
	@cd services/host_task && go run cmd/main.go &
	@cd services/auto_healing && go run cmd/main.go &
	@cd services/host_manager && go run cmd/main.go &
	@cd services/scenario && go run cmd/main.go &
	@cd services/audio_logs && go run cmd/main.go &
	@cd apps/gwall-e-agent && go run cmd/main.go &

test:
	@echo "Testing all services..."
	@cd services/host_task && go test ./...
	@cd services/auto_healing && go test ./...
	@cd services/host_manager && go test ./...
	@cd services/scenario && go test ./...
	@cd services/audio_logs && go test ./...