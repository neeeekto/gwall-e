# Конвенции тестов gwall-e

Канон **конвенций тестов** репозитория: фреймворк, suite-бутстрап, структура спеков,
мокинг портов. Тест-фреймворк — **Ginkgo v2 + Gomega** (версии pinned в `pkg/go.mod` —
здесь номера не дублируются как жёсткий факт). Команды прогона тестов — канон в
[build.md](build.md) (`cd pkg && go test ./...`); здесь они **не** повторяются — ссылка,
а не копия.

Все правила следуют стандарту [authoring.md](authoring.md) (MUST/SHOULD/WON'T, парность
«запрет → do», pointer-over-copy). Механизируемые правила несут пред-пометку будущего
enforcement-статуса (Phase 4 переключит её на фактическую — D-11, не ретрофит с нуля).

## Каркас suite (MUST)

Минимальный реальный эталон ниже **компилируется** в репозитории — это **не**
иллюстративный плейсхолдер.

- **MUST** бутстрапить suite через `RegisterFailHandler(Fail)` + `RunSpecs(t, "<Suite
  Name>")` внутри обычной `TestXxx(t *testing.T)`. Изобретать свой раннер — **WON'T**,
  потому что Ginkgo даёт готовую интеграцию с `go test` и репорт; вместо этого —
  стандартный бутстрап ниже. ⟶ convention-only (review-enforced)
- **MUST** использовать dot-imports `. "github.com/onsi/ginkgo/v2"` и
  `. "github.com/onsi/gomega"` — это идиома Ginkgo/Gomega (DSL `Describe`/`It`/`Expect`
  без префикса пакета). ⟶ convention-only (review-enforced)
- **MUST** держать тест-файлы `*_test.go` **рядом с кодом** — в том же пакете, либо во
  внешнем `<pkg>_test` (black-box). Складывать тесты в отдельное дерево вне пакета —
  **WON'T**, потому что Go-инструментарий ожидает `*_test.go` в каталоге пакета; вместо
  этого — файл рядом с тестируемым кодом. ⟶ convention-only (review-enforced)
- **MUST** писать комментарии в тестах **на английском** — канон языка см.
  [style.md](style.md); здесь правило не повторяется (единственное место — `style.md`).
  ⟶ convention-only (review-enforced)

```go
// Source: pkg/http/http_test.go — реальный, компилируется.
package http

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIdmServicesSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Package Suite")
}
```

## Структура спеков (SHOULD)

- **SHOULD** организовывать спеки иерархией `Describe` (юнит — что тестируем) →
  `Context` (сценарий — при каких условиях) → `It` (утверждение — что должно
  произойти). Плоский набор разрозненных `It` без группировки — **WON'T**, потому что
  теряется читаемость и общий setup; вместо этого — `Describe`/`Context` с `BeforeEach`.
  ⟶ convention-only (review-enforced)
- **SHOULD** использовать `DescribeTable` для табличных кейсов (один сценарий, много
  входов/выходов) вместо копипасты однотипных `It`. ⟶ convention-only (review-enforced)
- **MUST** ассертить через Gomega — `Expect(actual).To(matcher)` (и `ToNot`,
  `HaveOccurred`, `MatchError`, `Equal`, `BeAssignableToTypeOf` и т.п.). Голые
  `if got != want { t.Fatalf(...) }` в Ginkgo-спеке — **WON'T**, потому что мешают
  стили и теряют Gomega-репорт; вместо этого — matcher-ассерт Gomega.
  ⟶ convention-only (review-enforced)

Реальный пример спека (Source: `pkg/http/middlewares_test.go` — реальный, компилируется):

```go
var _ = Describe("CircuitBreakerMiddleware", func() {
	var (
		middleware  MiddlewareFunc
		nextHandler func(*http.Request) (*http.Response, error)
		req         *http.Request
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
		BeforeEach(func() {
			nextHandler = func(r *http.Request) (*http.Response, error) {
				// next returns a transport error; middleware must propagate it
				return nil, errors.New("connection error")
			}
		})

		It("should return the error", func() {
			resp, err := middleware(req, nextHandler)
			Expect(resp).To(BeNil())
			Expect(err).To(MatchError("connection error"))
		})
	})
})
```

## Мокинг портов (mockery + Gomega)

Тестирование use case'ов, зависящих от портов (`UnitOfWork`, repositories), —
**через генерируемые моки (кодоген)**, не вручную.

- **SHOULD** мокать порты через **mockery** (`vektra/mockery`) и проверять результат
  use case через Gomega-ассерт. Писать fake-реализации портов вручную — **WON'T**,
  потому что ручной фейк дрейфует от интерфейса при его изменении; вместо этого —
  генерируемый mockery-мок, синхронный с портом. ⟶ planned: Phase 4 (go:generate)
- **MUST** не выдавать mockery за «уже настроенный»: в репозитории его сейчас **нет**.
  Установка инструмента (`go install github.com/vektra/mockery/...`), `.mockery.yaml`
  и `go:generate`-обвязка — **planned: Phase 4** (ENF-05). Показывать рабочую команду
  генерации как проверенную — **WON'T** (no-phantom, [authoring.md](authoring.md));
  вместо этого фиксируется только **выбор инструмента и конвенция** ниже.

Целевой вид (иллюстрация — **не** код из компилируемого файла; плейсхолдер `Order`,
вне домена gwall-e). mockery v3 генерирует testify-style мок с expecter API
(`mock.EXPECT().Method().Return(...)`); конструктор `NewMockX(t)` авто-регистрирует
`t.Cleanup` с проверкой ожиданий, а в Ginkgo вместо `t` передаётся `GinkgoT()`. Матчеры
`mock.Anything` приходят из `github.com/stretchr/testify/mock` — это **обычный**
qualified-импорт, не dot-import (в отличие от ginkgo/gomega):

```go
// Иллюстрация / целевой вид (planned: Phase 4). НЕ из компилируемого файла.
import "github.com/stretchr/testify/mock" // обычный импорт; ginkgo/gomega — dot-import

var _ = Describe("RegisterOrderUseCase", func() {
	It("saves the order via the port", func() {
		repo := NewMockOrderRepository(GinkgoT()) // авто-Cleanup сверит ожидания
		repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

		uc := NewRegisterOrderUseCase(repo)
		out, err := uc.Execute(ctx, RegisterOrderInput{ /* ... */ })

		Expect(err).ToNot(HaveOccurred()) // результат проверяем через Gomega
		Expect(out.ID).ToNot(BeEmpty())
	})
})
```

## Что где живёт (без дублей)

- **Команды прогона** (`cd pkg && go test ./...`, `GOWORK=off` для `inventory`, версия
  Go) — канон в [build.md](build.md) (раскладка модулей — `structure.md`). Перечислять
  их здесь — **WON'T**; вместо этого — ссылка.
- **Язык комментариев в тестах** (английский) — канон в [style.md](style.md). `testing.md`
  ссылается за правилом, **не** формулирует его сам — **WON'T** заводить второй источник
  истины по языку (карта владения, `boundaries.md`).
