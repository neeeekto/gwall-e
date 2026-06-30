---
phase: 6
slug: inventory
status: ready
nyquist_compliant: true
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
| 06-01-T1 | 06-01 | 1 | INV-03 | T-06-01,T-06-02 | Типизированные ID-VO; parse→ErrInvalidID | unit (ID-VO) | cd services/inventory && go test ./internal/domain/... -run TestDomainSuite -count=1 | ❌ Wave 0 (this task) | ⬜ pending |
| 06-01-T2 | 06-01 | 1 | EVT-01,EVT-02 | T-06-03 | aggregateBase record/PullEvents; EventEnvelope+Actor форма | unit (aggregateBase) | cd services/inventory && go test ./internal/domain/... -count=1 | ❌ Wave 0 (this task) | ⬜ pending |
| 06-01-T3 | 06-01 | 1 | SVC-01,INV-08,INV-10 | T-06-04,T-06-SC | Порты + типизированные конфликты; mockery-моки | structural (build+mock gen) | cd services/inventory && make generate-mocks && go build ./... && go vet ./internal/domain/... | ❌ Wave 0 (this task) | ⬜ pending |
| 06-02-T1 | 06-02 | 2 | HW-01..06 | T-06-07 | Immutable HostHardware VO; defensive copy | unit (VO) DescribeTable | cd services/inventory && go test ./internal/domain/... -count=1 | ❌ → 06-02-T1 | ⬜ pending |
| 06-02-T2 | 06-02 | 2 | INV-02,03,04,05,06,07,09,LOC-03,EVT-01 | T-06-05,T-06-06,T-06-08 | Host фабрика + 3-state lifecycle SM; deleted≠enum | unit (state-machine) DescribeTable | cd services/inventory && go test ./internal/domain/... -count=1 | ❌ → 06-02-T2 | ⬜ pending |
| 06-03-T1 | 06-03 | 2 | INV-01,03,09,EVT-01 | T-06-10,T-06-11 | Project фабрика; Owner raw string; события | unit (aggregate) | cd services/inventory && go test ./internal/domain/... -count=1 | ❌ → 06-03-T1 | ⬜ pending |
| 06-03-T2 | 06-03 | 2 | LOC-01,02,04,EVT-01 | T-06-09,T-06-11 | DC/Module/Rack 3 агрегата; иерархия по ID | unit (aggregate) | cd services/inventory && go test ./internal/domain/... -count=1 | ❌ → 06-03-T2 | ⬜ pending |
| 06-04-T1 | 06-04 | 3 | EVT-02,SVC-01 | T-06-13 | enrich envelope через Clock/IDGen-порты; фейки | usecase (fake clock/idgen) | cd services/inventory && go test ./internal/usecase/... -run TestUseCaseSuite -count=1 | ❌ Wave 0 (this task) | ⬜ pending |
| 06-04-T2 | 06-04 | 3 | INV-02,08,10,EVT-01,EVT-02 | T-06-12,T-06-15,T-06-16 | RegisterHost FQDN-конфликт+advisory; atomicity-форма | usecase (mock repo/uniq/advisor + fakes) | cd services/inventory && go test ./internal/usecase/... -count=1 | ❌ → 06-04-T2 | ⬜ pending |
| 06-04-T3 | 06-04 | 3 | INV-05,06,07,EVT-01 | T-06-14 | Decommission/Delete(repo.Delete)/Reassign/Relocate/ChangeHW | usecase (mock repo + fakes) | cd services/inventory && go test ./internal/usecase/... -count=1 && go vet ./internal/usecase/... | ❌ → 06-04-T3 | ⬜ pending |
| 06-05-T1 | 06-05 | 4 | INV-01,10,EVT-01,EVT-02 | T-06-17,T-06-18 | CreateProject; DeleteProject delete-if-empty query-порт | usecase (mock repo/hostCount + fakes) | cd services/inventory && go test ./internal/usecase/... -count=1 | ❌ → 06-05-T1 | ⬜ pending |
| 06-05-T2 | 06-05 | 4 | LOC-01,02,EVT-01,SVC-01 | T-06-19 | DC/Module/Rack CRUD usecases; полный suite | usecase (mock repos + fakes) | cd services/inventory && go test ./... -count=1 && go vet ./... | ❌ → 06-05-T2 | ⬜ pending |
| 06-06-T1 | 06-06 | 4 | DOC-07 | T-06-20,T-06-21 | glossary.md ubiquitous language; имена событий | manual review + grep | test -f knowledge/glossary.md && grep -qi decommission knowledge/glossary.md && grep -q HostRegistered knowledge/glossary.md | ❌ → 06-06-T1 | ⬜ pending |
| 06-06-T2 | 06-06 | 4 | SVC-01 | — | README индекс glossary; канон-слои build | structural (grep + build) | grep -A1 glossary.md knowledge/README.md | grep -qi существует; cd services/inventory && go build ./... && go vet ./... | ❌ → 06-06-T2 | ⬜ pending |

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

**Approval:** approved (planner — all tasks carry <automated> verify; Wave-0 infra folded into 06-01 T1/T3 + 06-04 T1)
