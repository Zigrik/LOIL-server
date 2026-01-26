package world

import (
	"LOIL-server/internal/config"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// IntSlice - срез int с кастомной JSON сериализацией
type IntSlice []int

// MarshalJSON преобразует IntSlice в строку для JSON
func (is IntSlice) MarshalJSON() ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("\"")
	for i, val := range is {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%d", val))
	}
	sb.WriteString("\"")
	return []byte(sb.String()), nil
}

// UnmarshalJSON преобразует строку из JSON в IntSlice
func (is *IntSlice) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	parts := strings.Fields(str)
	*is = make(IntSlice, len(parts))

	for i, part := range parts {
		var val int
		if _, err := fmt.Sscanf(part, "%d", &val); err != nil {
			return fmt.Errorf("ошибка парсинга значения '%s': %v", part, err)
		}
		(*is)[i] = val
	}

	return nil
}

// WorldObject - игровой объект на карте
type WorldObject struct {
	ID          int                    `json:"id"`
	TypeID      int                    `json:"type_id"`      // ID из конфига
	X           int                    `json:"x"`            // Позиция в локации
	Y           int                    `json:"y"`            // Для будущей 2D реализации
	LocationID  int                    `json:"location_id"`  // ID локации
	Durability  int                    `json:"durability"`   // Текущая прочность
	GrowthStage int                    `json:"growth_stage"` // Стадия роста (0-100)
	Storage     map[int]int            `json:"storage"`      // ID предмета -> количество
	CustomData  map[string]interface{} `json:"custom_data"`  // Дополнительные данные
}

// InventoryItem - предмет в инвентаре
type InventoryItem struct {
	ItemID int `json:"item_id"` // ID типа предмета из конфига
	Count  int `json:"count"`   // Количество
}

// Character - персонаж
type Character struct {
	ID         int                   `json:"id"`
	Name       string                `json:"name"`
	Location   int                   `json:"location"`
	X          float64               `json:"x"`
	Speed      float64               `json:"speed"`
	Direction  int                   `json:"direction"`
	Controlled int                   `json:"controlled"`
	Vertical   int                   `json:"-"`
	Inventory  map[int]InventoryItem `json:"inventory"`  // ID слота -> предмет
	Equipped   map[string]int        `json:"equipped"`   // Тип инструмента -> item_id
	HandsFree  bool                  `json:"hands_free"` // Руки свободны
}

// Location - локация мира
type Location struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Foreground  IntSlice               `json:"foreground"` // ID объектов переднего плана
	Road        IntSlice               `json:"road"`       // ID типов дороги
	Ground      IntSlice               `json:"ground"`     // ID типов земли
	Background  IntSlice               `json:"background"` // ID объектов заднего плана
	Objects     map[int]*WorldObject   `json:"objects"`    // Дополнительные объекты (ключ - позиция)
	Transitions map[string]*Transition `json:"transitions"`
}

// Transition - переход между локациями
type Transition struct {
	LocationID int    `json:"location_id"`
	Type       string `json:"type"`
}

// Обновим структуру CreatureBehavior
type CreatureBehavior struct {
	Type             string    `json:"type"`                // wander, eat, rest, attack, flee
	TargetPos        int       `json:"target_pos"`          // Целевая позиция
	Duration         float64   `json:"duration"`            // Длительность поведения в секундах
	StartTime        time.Time `json:"start_time"`          // Время начала поведения
	Cooldown         float64   `json:"cooldown"`            // Время перезарядки
	AteAtCurrentStop bool      `json:"ate_at_current_stop"` // Уже поел на этой остановке
}

// Creature - существо (NPC)
type Creature struct {
	ID              int                   `json:"id"`
	TypeID          int                   `json:"type_id"`     // ID из конфига
	Name            string                `json:"name"`        // Имя (если есть)
	Location        int                   `json:"location"`    // ID локации
	X               float64               `json:"x"`           // Позиция
	Health          int                   `json:"health"`      // Текущее здоровье
	MaxHealth       int                   `json:"max_health"`  // Максимальное здоровье
	Hunger          int                   `json:"hunger"`      // Голод (0-100)
	Thirst          int                   `json:"thirst"`      // Жажда (0-100)
	CurrentBehavior *CreatureBehavior     `json:"behavior"`    // Текущее поведение
	Inventory       map[int]InventoryItem `json:"inventory"`   // Инвентарь
	LastUpdate      time.Time             `json:"last_update"` // Время последнего обновления
}

// Interaction - базовая структура взаимодействия (для совместимости)
// Более полная версия в config пакете
type Interaction struct {
	Type string `json:"type"`
	Tool string `json:"tool"`
	Time int    `json:"time"`
}

// Обновим структуру World
type World struct {
	PlayerID   int                  `json:"player_id"`
	Characters []*Character         `json:"characters"`
	Locations  []*Location          `json:"locations"`
	Objects    map[int]*WorldObject `json:"objects"`   // Все объекты мира
	Creatures  []*Creature          `json:"creatures"` // Все существа мира
	Configs    *config.Configs      `json:"-"`         // Конфигурации (не сериализуется в JSON)
}
