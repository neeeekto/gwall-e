---
name: write-tests
description: "Написать тесты для Go-кода в gwall-e по правилам проекта (Ginkgo + Gomega). Используй, когда просят покрыть тестами хендлер, агрегат, value object, репозиторий или пакет pkg/. Процедура: правила берутся из memory-bank/testing.md, скила добавляет рабочий процесс и правильные команды прогона."
trigger: /write-tests
---

# /write-tests

Пишет тесты для Go-кода gwall-e по конвенциям проекта.

> **Источник истины по правилам — [`memory-bank/testing.md`](../../../memory-bank/testing.md).**
> Скила его не дублирует. Если правила в memory-bank изменились — следуй им, а не этому файлу.

## Использование

```
/write-tests <путь к файлу/пакету>   # покрыть конкретный код
/write-tests                          # покрыть код из текущего контекста (выделение/обсуждаемый файл)
```

## Процедура

1. **Прочитай правила.** Открой [`memory-bank/testing.md`](../../../memory-bank/testing.md). Кратко: Ginkgo (`Describe`/`Context`/`It`) + Gomega (`Expect(...).To(...)`); **комментарии и строки спецификаций — на английском** (исключение из общего правила проекта).

2. **Пойми код под тест.** Прочитай целевой файл и его соседей. Для `services/inventory` держи в голове архитектуру (агрегаты, хендлеры команд/запросов, порты) — см. корневой [`CLAUDE.md`](../../../CLAUDE.md).
   - Агрегат → проверяй инварианты фабрики `NewX(...)`, накопление и `PullEvents()`.
   - Хендлер команды/запроса → мокай порты (репозитории, паблишеры), проверяй маппинг DTO → домен и возвращаемые ошибки.
   - Sentinel-ошибки домена сверяй через `Expect(err).To(MatchError(domain.ErrXxx))`.

3. **Заведи suite-файл, если его нет.** Один на пакет, идиома проекта (см. `pkg/http/http_test.go`):
   ```go
   package <pkg>

   import (
       "testing"

       . "github.com/onsi/ginkgo/v2"
       . "github.com/onsi/gomega"
   )

   func Test<Pkg>Suite(t *testing.T) {
       RegisterFailHandler(Fail)
       RunSpecs(t, "<Human Readable> Suite")
   }
   ```

4. **Пиши спеки.** Структура BDD; happy path и граничные/ошибочные случаи. Комментарии и описания — на английском:
   ```go
   var _ = Describe("HostRegistry", func() {
       // happy path: a new host is stored and retrievable
       It("registers a new host", func() {
           Expect(registry.Register(host)).To(Succeed())
       })
   })
   ```

5. **Прогоняй правильной командой** (см. [`CLAUDE.md`](../../../CLAUDE.md), раздел «Сборка и тесты»):
   - `pkg/` (в workspace): `cd pkg && go test ./...`
   - `services/inventory` (**НЕ в `go.work`**): `cd services/inventory && GOWORK=off go test ./...`
     Без `GOWORK=off` Go ругнётся, что каталог не в workspace.
   - Итерируй, пока зелёные.

## Не делай

- Не переписывай существующие тесты на стандартном `testing` ради миграции на Ginkgo — только новый код (правило из `memory-bank/testing.md`).
- Не «чини» попутно `nil`-зависимости и `// TODO` в `inventory/cmd/main.go` — это строительные леса, а не баги.
- Не пиши комментарии в тестах по-русски — здесь действует исключение: английский.
