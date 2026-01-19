package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// InteractionResult - результат взаимодействия
type InteractionResult struct {
	ItemID int `json:"item_id"`
	Count  int `json:"count"`
}

// Interaction - взаимодействие с объектом
type Interaction struct {
	Type    string              `json:"type"`    // pick, chop, mine, harvest, collect
	Tool    string              `json:"tool"`    // hand, axe, pickaxe, knife
	Time    int                 `json:"time"`    // Время в секундах
	Results []InteractionResult `json:"results"` // Результаты
}

// ObjectTypeConfig - конфигурация типа объекта
type ObjectTypeConfig struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Layer         string        `json:"layer"` // foreground, background, road_level
	Size          int           `json:"size"`  // 1, 2, 3
	MaxDurability int           `json:"max_durability"`
	GrowthTime    int           `json:"growth_time"` // 0 для нерастущих
	Interactions  []Interaction `json:"interactions"`
}

// RoadTypeConfig - конфигурация типа дороги
type RoadTypeConfig struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Durability  int     `json:"durability"`
	SpeedMod    float64 `json:"speed_mod"`
}

// GroundTypeConfig - конфигурация типа земли
type GroundTypeConfig struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Walkable    bool   `json:"walkable"`
	Buildable   bool   `json:"buildable"`
	ResourceID  int    `json:"resource_id"`
}

// Configs - все конфигурации
type Configs struct {
	ObjectTypes map[string]*ObjectTypeConfig `json:"object_types"`
	RoadTypes   map[string]*RoadTypeConfig   `json:"road_types"`
	GroundTypes map[string]*GroundTypeConfig `json:"ground_types"`
}

// LoadConfigs загружает все конфигурации
func LoadConfigs() (*Configs, error) {
	configs := &Configs{
		ObjectTypes: make(map[string]*ObjectTypeConfig),
		RoadTypes:   make(map[string]*RoadTypeConfig),
		GroundTypes: make(map[string]*GroundTypeConfig),
	}

	// Определяем путь к конфигурациям
	configDir := "internal/config"

	// Загружаем типы объектов
	if err := loadJSON(filepath.Join(configDir, "object_types.json"), &configs.ObjectTypes); err != nil {
		return nil, err
	}

	// Загружаем типы дорог
	if err := loadJSON(filepath.Join(configDir, "road_types.json"), &configs.RoadTypes); err != nil {
		return nil, err
	}

	// Загружаем типы земли
	if err := loadJSON(filepath.Join(configDir, "ground_types.json"), &configs.GroundTypes); err != nil {
		return nil, err
	}

	return configs, nil
}

func loadJSON(filename string, target interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
