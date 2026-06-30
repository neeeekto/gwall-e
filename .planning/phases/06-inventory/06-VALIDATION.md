---
phase: 6
slug: inventory
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-06-30
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test + Ginkgo v2 + Gomega (mockery v3 for port mocks) |
| **Config file** | `services/inventory/.mockery.yaml` (extend in Wave 0 with real domain ports) |
| **Quick run command** | `cd services/inventory && go test ./internal/domain/... ./internal/usecases/...` |
| **Full suite command** | `cd services/inventory && go test ./...` |
| **Estimated runtime** | ~10 seconds (pure unit, no containers) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/domain/... ./internal/usecases/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** ~10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (filled by planner) | — | — | — | — | — | — | — | — | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

*Note: aggregates/invariants verified by direct unit tests (pure functions/factories); usecases verified on in-memory fakes + mockery mocks for ports (D-03).*

---

## Wave 0 Requirements

- [ ] Extend `services/inventory/.mockery.yaml` with real domain ports (repo / UnitOfWork / Outbox / uniqueness query-port)
- [ ] In-memory fake helpers for UnitOfWork (calls `fn`) + Outbox (slice) for usecase tests
- [ ] Ginkgo suite bootstrap files for `internal/domain` and `internal/usecases`

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `knowledge/glossary.md` captures ubiquitous language | DOC-07 | Prose/doc artifact, not executable | Review glossary contains Project/Host/Owner/Module/Connection, identity, `decommission ≠ delete`, semantic event names |

*Remaining phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
