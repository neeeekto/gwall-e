---
phase: 4
slug: enforcement
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-06-17
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test + Ginkgo v2/Gomega (existing in `pkg`); config-self-verification (`golangci-lint config verify`, `buf build`, `lefthook validate`) |
| **Config file** | none — Wave 0 installs (`.golangci.yml`, `lefthook.yml`, commitlint, buf) are the artifacts under test |
| **Quick run command** | `cd pkg && go test ./...` |
| **Full suite command** | `go test ./... && golangci-lint config verify && buf build` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `{quick run command}`
- **After every plan wave:** Run `{full suite command}`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| {N}-01-01 | 01 | 1 | ENF-{XX} | T-4-01 / — | {expected secure behavior or "N/A"} | config-verify | `{command}` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

*Filled by planner/executor — each ENF artifact gets a self-verification command (`golangci-lint config verify`, `lefthook validate`, `npx commitlint`, `buf build`).*

---

## Wave 0 Requirements

- [ ] `.golangci.yml` — config under test for ENF-01 (verify via `golangci-lint config verify`)
- [ ] `lefthook.yml` — config under test for ENF-02 (verify via `lefthook validate`)
- [ ] commitlint config + `package.json` — ENF-03 (verify via `echo 'feat: x' | npx --no-install commitlint`)
- [ ] `buf.yaml` + `buf.gen.yaml` — ENF-04 skeleton (verify via `buf build` / `buf lint` on empty module)
- [ ] tool install bootstrap (`make tools`) — golangci-lint/gofumpt/gci/lefthook/buf absent on machine

*ENF-05 (status flips in `knowledge/*.md`) is doc-level — verified by grep that no stale `planned:` marks remain and statuses match the legend.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Git hooks actually fire on commit/push | ENF-02 | Requires a real git operation + installed tools | After `lefthook install` + `make tools`: make a bad commit msg → commit-msg hook rejects; introduce lint error → pre-commit rejects |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
