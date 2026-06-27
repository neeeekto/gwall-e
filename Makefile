# Source: tool versions verified vs official release pages this session — CITED (D-11).
# Reproducible pinned install of Go dev tools. NO @latest — exact versions only.
# NOTE: the stricter gofmt formatter runs EMBEDDED inside golangci-lint v2 formatters,
#       so it is NOT installed standalone here (standalone build requires Go 1.25; repo is Go 1.24.6).
# This Makefile is the chosen pinning mechanism (D-11): the root go.mod is a rotten
# leftover (stale ginkgo v1 / mongo v1, outside go.work) — do NOT add a `tool` block there.

GOLANGCI_VERSION := v2.12.2
LEFTHOOK_VERSION := v2.1.9
BUF_VERSION      := v1.71.0
MOCKERY_VERSION := v3.7.1

.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)
	go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
	@echo "Next: 'npm install' (commitlint) and 'lefthook install' (git hooks)."

# --- Dev stand & test targets (D-09/D-15/SVC-06) ---
# Use the docker compose v2 plugin (NOT the legacy docker-compose binary).

.PHONY: dev-up
dev-up:
	docker compose up -d

# Provision Kafka bootstrap topics via the inventory bootstrap CLI (CLI lands in Plan 04).
.PHONY: topics
topics:
	cd services/inventory && KAFKA_BROKERS=localhost:9092 go run ./cmd

# Integration tests gated behind the `integration` build tag (D-15) — Docker required.
.PHONY: test-integration
test-integration:
	cd services/inventory && go test -tags=integration ./...

# Regenerate mocks via the pinned mockery binary (installed by `make tools`).
.PHONY: generate-mocks
generate-mocks:
	mockery
