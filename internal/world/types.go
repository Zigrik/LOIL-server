package world

import (
	"LOIL-server/internal/config"
	"encoding/json"
	"fmt"
	"strings"
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

// Character - персонаж
type Character struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Location   int     `json:"location"`
	X          float64 `json:"x"`
	Speed      float64 `json:"speed"`
	Direction  int     `json:"direction"`
	Controlled int     `json:"controlled"`
	Vertical   int     `json:"-"`
}

// World - игровой мир
type World struct {
	PlayerID   int                  `json:"player_id"`
	Characters []*Character         `json:"characters"`
	Locations  []*Location          `json:"locations"`
	Objects    map[int]*WorldObject `json:"objects"` // Все объекты мира
	Configs    *config.Configs      `json:"-"`       // Конфигурации (не сериализуется в JSON)
}
