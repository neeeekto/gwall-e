---
phase: 03-conventions-architecture-docs
reviewed: 2026-06-17T17:10:00Z
depth: standard
files_reviewed: 7
files_reviewed_list:
  - knowledge/README.md
  - knowledge/architecture.md
  - knowledge/boundaries.md
  - knowledge/patterns.md
  - knowledge/style.md
  - knowledge/testing.md
  - AGENTS.md
findings:
  critical: 1
  warning: 4
  info: 2
  total: 7
status: issues_found
---

# Phase 03: Code Review Report (Conventions / Architecture Docs)

**Reviewed:** 2026-06-17T17:10:00Z
**Depth:** standard
**Files Reviewed:** 7
**Status:** issues_found

## Summary

Seven Markdown knowledge-base docs reviewed against documentation-appropriate criteria
(no-phantom, cross-reference integrity, rule-ownership/no-duplication, Go snippet
correctness, authoring-standard compliance). Russian-prose / English-code-term mixing was
explicitly out of scope and not flagged.

Overall the docs are well-structured: all relative markdown links resolve to existing
files, the fact-ownership map (boundaries.md) is consistent, the code-comment-language rule
correctly lives only in style.md with testing.md pointing to it, and architecture rules are
canonical in architecture.md with patterns.md referencing rather than copying them.

However the adversarial pass surfaced one no-phantom violation (the highest-priority defect
class for this project) plus several snippet-correctness and staleness defects:

- **CR-01 (BLOCKER):** testing.md presents a hand-edited, **non-compiling** code block as a
  verbatim "Source: ... компилируется" excerpt of `pkg/http/middlewares_test.go`. The shown
  code does not match the real file and would not compile — a direct no-phantom violation.
- **WR-01:** the shared `Order` placeholder factory has contradictory signatures across the
  two docs that explicitly claim to reuse it (`NewOrder` returns `(*Order, error)` in
  patterns.md but is called as a single-value return in architecture.md → won't compile).
- **WR-02 / WR-03 / WR-04:** stale "planned, Phase 3" tags pointing at docs that now exist;
  a `mock.`-qualified call with no shown import in the mockery snippet; and a missing
  forward markdown link for cross-refs to existing files.

## Critical Issues

### CR-01: testing.md presents non-compiling, altered code as a verbatim compiling source excerpt (no-phantom)

**File:** `knowledge/testing.md:65-83`
**Issue:** The spec block is introduced as
`Реальный пример спека (Source: pkg/http/middlewares_test.go, компилируется)`, asserting it
is verbatim, real, and compiling. It is none of those:

- The real `middlewares_test.go` `BeforeEach` (lines 20-29) builds a full config
  `CircuitBreakerConfig{MaxRequests: 1, Interval: 1*time.Second, Timeout: 1*time.Second,
  MaxFailures: 1}` and constructs `req, _ = http.NewRequest(...)`. The doc snippet shows
  only `CircuitBreakerConfig{MaxFailures: 1}` and declares no `req`.
- The doc snippet references `failingHandler`, an identifier that **does not exist** in the
  real file (the real test uses locally-defined `nextHandler` closures).
- As written, the doc snippet would not compile: `req` is undeclared and `failingHandler`
  is undefined.

This is exactly the no-phantom failure the project flags as highest priority: documenting a
code artifact (a "compiling source excerpt") that does not exist as shown. The same `Source:
... компилируется` claim on the suite snippet (lines 33-48) IS accurate against
`pkg/http/http_test.go`, which makes the false claim here more misleading by association.

**Fix:** Either (a) paste the real excerpt verbatim and keep the "компилируется" claim, or
(b) drop the `Source:`/"компилируется" claim and label it illustrative. Minimal real-faithful
version:

```go
// Source: pkg/http/middlewares_test.go — реальный, компилируется.
var _ = Describe("CircuitBreakerMiddleware", func() {
    var (
        middleware MiddlewareFunc
        req        *http.Request
    )

    BeforeEach(func() {
        config := CircuitBreakerConfig{
            MaxRequests: 1,
            Interval:    1 * time.Second,
            Timeout:     1 * time.Second,
            MaxFailures: 1,
        }
        middleware = CircuitBreakerMiddleware(config)
        req, _ = http.NewRequest("GET", "http://example.com", nil)
    })

    Context("when next handler returns regular error", func() {
        It("should return the error", func() {
            nextHandler := func(r *http.Request) (*http.Response, error) {
                return nil, errors.New("connection error")
            }
            resp, err := middleware(req, nextHandler)
            Expect(resp).To(BeNil())
            Expect(err).To(MatchError("connection error"))
        })
    })
})
```

## Warnings

### WR-01: `NewOrder` factory signature contradicts across architecture.md and patterns.md (won't compile)

**File:** `knowledge/architecture.md:78` (vs `knowledge/patterns.md:53,132`)
**Issue:** Both docs state the `Order` placeholder is deliberately shared/reused between
them (architecture.md:16, patterns.md:16). patterns.md defines the factory as
`func NewOrder(sku string, qty int) (*Order, error)` (line 132) and correctly consumes it
as `order, err := NewOrder(in.SKU, in.Qty)` (line 53). But architecture.md:78 calls the
same factory as a single-return:

```go
order := NewOrder(in.SKU, in.Qty) // фабрика агрегата держит инварианты
```

Against the shared signature this discards the `error` return and would not compile. Beyond
the compile issue, it undercuts the doc's own §"Доменные события" rule that the factory
"держит инварианты" — a factory that enforces invariants must be able to return an error,
which architecture.md's call site silently ignores.

**Fix:** Make architecture.md:76-82 consistent with the canonical factory signature:

```go
order, err := NewOrder(in.SKU, in.Qty) // фабрика агрегата держит инварианты
if err != nil {
    return err
}
// запись агрегата + outbox-событий — внутри той же tx
return saveOrderAndOutbox(ctx, order)
```

### WR-02: Stale "planned, Phase 3" tags on docs that already exist (inverse-phantom / staleness)

**File:** `knowledge/architecture.md:6`; `knowledge/style.md:36`; `knowledge/style.md:108`
**Issue:** These cross-references describe sibling docs as not-yet-created:

- architecture.md:6 — "в `patterns.md` (planned, Phase 3)"
- style.md:36 — "Полная конвенция тестов — `testing.md` (planned, Phase 3)"
- style.md:108 — "`architecture.md` (planned, Phase 3)"

All three target files exist now (`knowledge/patterns.md`, `knowledge/testing.md`,
`knowledge/architecture.md`) and are listed as `существует` in both README.md and AGENTS.md
indexes. Per the project's own no-phantom inverse (README.md:37-38, authoring.md:61-62):
existing files are referenced with a markdown link, only future docs are mentioned link-less
with a "запланировано" status. These lines have it backwards — they describe existing files
as planned and render them as bare backtick names instead of links, contradicting the index.

**Fix:** Convert each to a live relative markdown link and drop the "(planned, Phase 3)"
qualifier, e.g. architecture.md:6 → "пошаговые рецепты … — в [patterns.md](patterns.md),
который ссылается сюда"; style.md:36 → "Полная конвенция тестов — [testing.md](testing.md)";
style.md:108 → "канон в [architecture.md](architecture.md)".

### WR-03: mockery snippet uses `mock.Anything` with no shown/declared import

**File:** `knowledge/testing.md:111`
**Issue:** The illustrative spec uses `mock.Anything` (testify's matcher package, line 111)
but the snippet and its surrounding suite only establish dot-imports for ginkgo/gomega
(lines 22, 37-42). `mock` is a regular qualified import (`github.com/stretchr/testify/mock`)
and is neither imported nor mentioned. The doc otherwise goes to lengths to assert API
fidelity ("mockery v3 генерирует testify-style мок с expecter API", line 102-104), so an
unqualified/undeclared `mock.` reference undermines that fidelity claim and would not
compile if a reader copies it as the "целевой вид".

**Fix:** Add a one-line note that the expecter matchers come from
`github.com/stretchr/testify/mock` (imported normally, not dot-imported), or qualify the
import in the snippet header comment so the placeholder is copy-faithful.

### WR-04: §"MUST NOT CQRS" bullets rely on a trailing shared enforcement label rather than per-rule tags

**File:** `knowledge/architecture.md:141-148`
**Issue:** authoring.md §"Статус enforcement" says mechanizable rules SHOULD carry an
enforcement label. The two `WON'T` bullets at lines 142-146 (no CQRS dispatcher; no
`TxManager`) carry no inline label; instead a single shared `⟶ planned: CI-gated Phase 4`
sits at line 148 after the block. This is defensible as a section-level label, but it is
inconsistent with every other rule in the file (which carries its own inline arrow) and is
fragile: if a third bullet is inserted before line 148 it will silently inherit the label.
Counts confirm the gap — 12 normative bullets vs 9 inline enforcement arrows in this file.

**Fix:** Attach `⟶ planned: CI-gated Phase 4 (depguard)` inline to each of the two `WON'T`
bullets, or add an explicit "оба правила выше:" lead-in to line 148 so the shared scope is
unambiguous.

## Info

### IN-01: boundaries.md normative rules carry no enforcement labels

**File:** `knowledge/boundaries.md:13-21, 25-29, 33-37, 41-44, 47-53, 72-74`
**Issue:** All 6 normative MUST/WON'T bullets in boundaries.md lack an enforcement-status
label (0 arrows across the file). Most boundaries rules are non-mechanizable do-not
guidance, for which authoring.md's "where rule is mechanizable" carve-out applies — so this
is not a violation. But at least two are arguably mechanizable and could carry a label for
consistency: "WON'T ре-раздувать корневой CLAUDE.md … через generate-claude-md" (47-53) and
the fact-ownership "MUST ссылаться … относительной markdown-ссылкой … Копировать … WON'T"
(72-74). Consider `convention-only (review-enforced)` tags there.
**Fix:** Add `⟶ convention-only (review-enforced)` to the two mechanizable bullets; leave
the genuinely non-mechanizable do-not rules untagged.

### IN-02: `failingHandler` ghost identifier also weakens the §"Структура спеков" rule example

**File:** `knowledge/testing.md:77-81`
**Issue:** Secondary to CR-01: the same fabricated `failingHandler` is what the inline
comment ("next returns a transport error; middleware must propagate it") describes, so the
illustrative value of the §"Структура спеков" example partly rests on an identifier that
does not exist in the cited source. Once CR-01 is fixed with the real excerpt this resolves
automatically; noted separately so the comment/identifier pairing is rechecked during the
fix.
**Fix:** Covered by CR-01 — ensure the replacement excerpt's inline comment matches a real
handler/closure in the file.

---

## Verification notes (what was checked and passed)

- All relative markdown links in the 7 reviewed files resolve to existing targets (scripted
  link check, 0 broken).
- `knowledge/` contains all 10 docs the README/AGENTS indexes mark as `существует`; no index
  entry links to a missing file; future docs (`glossary.md`) correctly listed link-less.
- `pkg/http/http_test.go` matches testing.md's suite snippet (lines 33-48) verbatim — that
  "компилируется" claim is accurate.
- Code-comment-language rule lives only in style.md (§"Язык кода и комментариев");
  testing.md:29-31 and :127-129 correctly point to it without restating it.
- Architecture rules are canonical in architecture.md; patterns.md references them per
  section without duplicating (pointer-over-copy honored), except the WR-01 signature drift.
- `03-VALIDATION.md` referenced at architecture.md:56 exists at the cited path.
- Platform security controls (ownership races, SSH rights, audit) are explicitly marked
  non-existent/convention, not documented as live (architecture.md:134-136) — no-phantom
  honored there.

---

_Reviewed: 2026-06-17T17:10:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
