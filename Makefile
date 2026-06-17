# Source: tool versions verified vs official release pages this session — CITED (D-11).
# Reproducible pinned install of Go dev tools. NO @latest — exact versions only.
# NOTE: the stricter gofmt formatter runs EMBEDDED inside golangci-lint v2 formatters,
#       so it is NOT installed standalone here (standalone build requires Go 1.25; repo is Go 1.24.6).
# This Makefile is the chosen pinning mechanism (D-11): the root go.mod is a rotten
# leftover (stale ginkgo v1 / mongo v1, outside go.work) — do NOT add a `tool` block there.

GOLANGCI_VERSION := v2.12.2
LEFTHOOK_VERSION := v2.1.9
BUF_VERSION      := v1.71.0

.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)
	go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	@echo "Next: 'npm install' (commitlint) and 'lefthook install' (git hooks)."
