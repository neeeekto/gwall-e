package domain

// Wave-1 placeholder-агрегаты (06-01, deviation Rule 3 — forward-dependency).
//
// Доменные порты (ports.go) ссылаются на агрегаты Host/Project/DC/Module/Rack, но сами
// агрегаты рождаются позже: Host — в 06-02, Project/DC/Module/Rack — в 06-03. Чтобы ядро
// домена компилировалось в Wave 1 (план требует `go build ./...` exit 0 уже сейчас),
// здесь объявлены МИНИМАЛЬНЫЕ заглушки: каждая встраивает aggregateBase и несёт свой
// типизированный ID. Все остальные поля, фабрики и переходы (lifecycle, state-machine,
// события) добавляются 06-02/06-03 — те планы РАСШИРЯЮТ эти типы (добавляют поля/методы
// в host.go/project.go/...), а НЕ переобъявляют `type Host struct` (иначе redeclaration).
//
// Файл — единственная точка объявления struct-агрегатов; при наполнении в 06-02/06-03
// эти заглушки сливаются с реальными определениями (поля/методы — в host.go/project.go и
// т.п., type-объявление остаётся здесь до явного решения executor'а соответствующего плана).

// Host-агрегат вынесен в host.go (06-02): полная фабрика, lifecycle state-machine и
// операции-события. Заглушка удалена — реальный тип живёт в host.go (иначе redeclaration).

// DC — агрегат дата-центра (наполняется 06-03).
type DC struct {
	aggregateBase
	id DCID
}

// ID возвращает идентификатор дата-центра.
func (d *DC) ID() DCID { return d.id }

// Module — агрегат модуля/зала (наполняется 06-03).
type Module struct {
	aggregateBase
	id ModuleID
}

// ID возвращает идентификатор модуля.
func (m *Module) ID() ModuleID { return m.id }

// Rack — агрегат стойки (наполняется 06-03).
type Rack struct {
	aggregateBase
	id RackID
}

// ID возвращает идентификатор стойки.
func (r *Rack) ID() RackID { return r.id }
