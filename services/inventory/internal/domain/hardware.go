package domain

import "fmt"

// HostHardware — единый immutable VO состава железа хоста (HW-01/D-07). Все поля приватны,
// мутации нет: изменение железа = собрать НОВЫЙ VO целиком через NewHostHardware и заменить
// его в агрегате (host.ChangeHardware) → ровно одно событие HostHardwareChanged (не per-компонент).
// Под-компоненты структурированы (RAM/CPU/Drive/NIC/PSU/StorageController/GPU/Chassis/Motherboard),
// а не плоские слайсы примитивов. Все внешние идентификаторы (serial/inv/MAC/model) — raw string
// (HW-06/D-05): домен не парсит и не типизирует их формат.
type HostHardware struct {
	name        string              // имя/обозначение конфигурации (HW-01)
	platform    string              // платформа/поколение шасси (HW-01)
	motherboard Motherboard         // паспорт материнской платы отдельно от шасси (HW-05)
	ipmiMAC     string              // MAC интерфейса IPMI/BMC (HW-01, raw string HW-06)
	ram         []RAMModule         // модули памяти (HW-02)
	cpu         []CPU               // процессоры (HW-02)
	drives      []Drive             // накопители (HW-02)
	nics        []NIC               // сетевые карты, структурированы (HW-03)
	psus        []PSU               // блоки питания — узлы power-зависимости (HW-04)
	storageCtl  []StorageController // контроллеры хранения/RAID (HW-05)
	gpus        []GPU               // внутренние GPU (HW-05)
	chassis     Chassis             // паспорт шасси отдельно от материнки (HW-05)
}

// Motherboard — паспорт материнской платы (HW-05). Внешние ID — raw string (HW-06).
type Motherboard struct {
	Model  string // модель платы
	Vendor string // производитель
	Serial string // серийный номер (raw string)
	Inv    string // инвентарный номер (raw string)
}

// RAMModule — модуль оперативной памяти (HW-02). Спеки + паспортные поля компонента.
type RAMModule struct {
	Slot     string // слот установки
	Model    string // модель
	Vendor   string // производитель
	Lot      string // партия (raw string)
	Serial   string // серийный номер (raw string)
	Inv      string // инвентарный номер (raw string)
	SizeGB   int    // объём, ГБ
	SpeedMHz int    // частота, МГц
}

// CPU — процессор (HW-02).
type CPU struct {
	Slot   string // сокет установки
	Model  string // модель
	Vendor string // производитель
	Lot    string // партия (raw string)
	Serial string // серийный номер (raw string)
	Inv    string // инвентарный номер (raw string)
	Cores  int    // число ядер
}

// Drive — накопитель (HW-02).
type Drive struct {
	Slot   string // слот/посадочное место
	Model  string // модель
	Vendor string // производитель
	Lot    string // партия (raw string)
	Serial string // серийный номер (raw string)
	Inv    string // инвентарный номер (raw string)
	SizeGB int    // объём, ГБ
	Type   string // тип носителя (SSD/HDD/NVMe — raw string)
}

// NIC — структурированный сетевой компонент (HW-03): модель/скорость/набор MAC'ов, НЕ
// плоский MACs[] на уровне хоста. MAC'и — raw string (HW-06).
type NIC struct {
	Model    string   // модель сетевой карты
	SpeedGbE int      // скорость линка, Гбит/с
	MACs     []string // MAC-адреса портов (raw string)
}

// PSU — блок питания (HW-04): узел power-зависимости.
type PSU struct {
	Model  string // модель
	Vendor string // производитель
	Serial string // серийный номер (raw string)
	Inv    string // инвентарный номер (raw string)
	PowerW int    // номинальная мощность, Вт
}

// StorageController — контроллер хранения/RAID (HW-05).
type StorageController struct {
	Model   string // модель контроллера
	Vendor  string // производитель
	Serial  string // серийный номер (raw string)
	Inv     string // инвентарный номер (raw string)
	RAIDLvl string // уровень RAID (raw string)
}

// GPU — внутренний графический ускоритель (HW-05).
type GPU struct {
	Model    string // модель
	Vendor   string // производитель
	Serial   string // серийный номер (raw string)
	Inv      string // инвентарный номер (raw string)
	MemoryGB int    // объём видеопамяти, ГБ
}

// Chassis — паспорт шасси отдельно от материнской платы (HW-05).
type Chassis struct {
	Model  string // модель шасси
	Vendor string // производитель
	Serial string // серийный номер (raw string)
	Inv    string // инвентарный номер (raw string)
	UnitsU int    // высота в юнитах (U)
}

// HardwareSpec — публичный входной тип для сборки HostHardware на edge (хендлер маппит
// DTO→спек). Все поля экспортируемы: спек собирается снаружи домена, валидируется и
// «застывает» в immutable VO конструктором NewHostHardware.
type HardwareSpec struct {
	Name        string
	Platform    string
	Motherboard Motherboard
	IPMIMac     string
	RAM         []RAMModule
	CPU         []CPU
	Drives      []Drive
	NICs        []NIC
	PSUs        []PSU
	StorageCtl  []StorageController
	GPUs        []GPU
	Chassis     Chassis
}

// NewHostHardware собирает immutable HostHardware из спека (HW-01…06/D-07). Точка простых
// инвариантов состава (напр. непустое name — V5 input validation). CRITICAL (Pitfall 2):
// слайсы Go — reference-типы, поэтому каждый слайс копируется защитно (defensive copy),
// иначе вызывающий мутирует «immutable» VO через сохранённую ссылку. NIC'и копируются
// глубоко — вложенный MACs[] тоже отвязывается от спека.
func NewHostHardware(spec HardwareSpec) (HostHardware, error) {
	if spec.Name == "" {
		return HostHardware{}, fmt.Errorf("hardware name is required: %w", ErrInvalidHardware)
	}

	return HostHardware{
		name:        spec.Name,
		platform:    spec.Platform,
		motherboard: spec.Motherboard,
		ipmiMAC:     spec.IPMIMac,
		ram:         append([]RAMModule(nil), spec.RAM...),
		cpu:         append([]CPU(nil), spec.CPU...),
		drives:      append([]Drive(nil), spec.Drives...),
		nics:        copyNICs(spec.NICs),
		psus:        append([]PSU(nil), spec.PSUs...),
		storageCtl:  append([]StorageController(nil), spec.StorageCtl...),
		gpus:        append([]GPU(nil), spec.GPUs...),
		chassis:     spec.Chassis,
	}, nil
}

// copyNICs делает глубокую копию слайса NIC: и внешний слайс, и вложенный MACs[] каждого
// NIC отвязываются от источника (Pitfall 2 — иначе мутация MACs протекает в VO).
func copyNICs(in []NIC) []NIC {
	if in == nil {
		return nil
	}
	out := make([]NIC, len(in))
	for i, n := range in {
		n.MACs = append([]string(nil), n.MACs...)
		out[i] = n
	}
	return out
}

// Name возвращает имя конфигурации железа (HW-01).
func (h HostHardware) Name() string { return h.name }

// Platform возвращает платформу (HW-01).
func (h HostHardware) Platform() string { return h.platform }

// Motherboard возвращает паспорт материнской платы (HW-05); value-копия, мутации не текут.
func (h HostHardware) Motherboard() Motherboard { return h.motherboard }

// IPMIMac возвращает MAC IPMI/BMC (HW-01, raw string HW-06).
func (h HostHardware) IPMIMac() string { return h.ipmiMAC }

// Chassis возвращает паспорт шасси (HW-05); value-копия.
func (h HostHardware) Chassis() Chassis { return h.chassis }

// RAM возвращает КОПИЮ слайса модулей памяти (Pitfall 2 — без копии VO протекает).
func (h HostHardware) RAM() []RAMModule { return append([]RAMModule(nil), h.ram...) }

// CPU возвращает КОПИЮ слайса процессоров (Pitfall 2).
func (h HostHardware) CPU() []CPU { return append([]CPU(nil), h.cpu...) }

// Drives возвращает КОПИЮ слайса накопителей (Pitfall 2).
func (h HostHardware) Drives() []Drive { return append([]Drive(nil), h.drives...) }

// NICs возвращает ГЛУБОКУЮ копию слайса сетевых карт — вложенный MACs[] тоже копируется
// (Pitfall 2 — иначе мутация MACs элемента протекает в VO).
func (h HostHardware) NICs() []NIC { return copyNICs(h.nics) }

// PSUs возвращает КОПИЮ слайса блоков питания (Pitfall 2).
func (h HostHardware) PSUs() []PSU { return append([]PSU(nil), h.psus...) }

// StorageControllers возвращает КОПИЮ слайса контроллеров хранения (Pitfall 2).
func (h HostHardware) StorageControllers() []StorageController {
	return append([]StorageController(nil), h.storageCtl...)
}

// GPUs возвращает КОПИЮ слайса GPU (Pitfall 2).
func (h HostHardware) GPUs() []GPU { return append([]GPU(nil), h.gpus...) }
