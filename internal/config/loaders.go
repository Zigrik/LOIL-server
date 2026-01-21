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
	Type              string              `json:"type"`                // pick, chop, mine, harvest, collect, dig
	Tool              string              `json:"tool"`                // hand, axe, pickaxe, knife, shovel
	Time              int                 `json:"time"`                // Время в секундах
	Results           []InteractionResult `json:"results"`             // Результаты
	ReduceDurability  int                 `json:"reduce_durability"`   // Сколько прочности отнимать (по умолчанию 1)
	TransformTo       int                 `json:"transform_to"`        // Во что превращается объект после взаимодействия
	DestroyOnComplete bool                `json:"destroy_on_complete"` // Уничтожать ли объект после взаимодействия
}

// ObjectTypeConfig - конфигурация типа объекта
type ObjectTypeConfig struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Foreground    bool          `json:"foreground"`
	RoadLevel     bool          `json:"road_level"`
	Background    bool          `json:"background"`
	Size          int           `json:"size"` // 1, 2, 3
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

// ItemTypeConfig - конфигурация типа предмета
type ItemTypeConfig struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"` // food, resource, tool, seed
	StackSize   int     `json:"stack_size"`
	Weight      float64 `json:"weight"`
}

// CreatureTypeConfig - конфигурация типа существа
type CreatureTypeConfig struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Type            string   `json:"type"` // humanoid, animal
	Size            int      `json:"size"`
	Health          int      `json:"health"`
	Damage          int      `json:"damage"`
	Speed           float64  `json:"speed"`
	FavoriteFoods   []int    `json:"favorite_foods"`
	Behaviors       []string `json:"behaviors"`
	DefaultBehavior string   `json:"default_behavior"`
}

// Configs - все конфигурации
type Configs struct {
	ObjectTypes   map[string]*ObjectTypeConfig   `json:"object_types"`
	RoadTypes     map[string]*RoadTypeConfig     `json:"road_types"`
	GroundTypes   map[string]*GroundTypeConfig   `json:"ground_types"`
	ItemTypes     map[string]*ItemTypeConfig     `json:"item_types"`
	CreatureTypes map[string]*CreatureTypeConfig `json:"creature_types"`
}

// LoadConfigs загружает все конфигурации
func LoadConfigs() (*Configs, error) {
	configs := &Configs{
		ObjectTypes:   make(map[string]*ObjectTypeConfig),
		RoadTypes:     make(map[string]*RoadTypeConfig),
		GroundTypes:   make(map[string]*GroundTypeConfig),
		ItemTypes:     make(map[string]*ItemTypeConfig),
		CreatureTypes: make(map[string]*CreatureTypeConfig),
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

	// Загружаем типы предметов
	itemTypesFile := filepath.Join(configDir, "item_types.json")
	if _, err := os.Stat(itemTypesFile); err == nil {
		if err := loadJSON(itemTypesFile, &configs.ItemTypes); err != nil {
			return nil, err
		}
	} else {
		configs.ItemTypes = make(map[string]*ItemTypeConfig)
	}

	// Загружаем типы существ
	creatureTypesFile := filepath.Join(configDir, "creature_types.json")
	if _, err := os.Stat(creatureTypesFile); err == nil {
		if err := loadJSON(creatureTypesFile, &configs.CreatureTypes); err != nil {
			return nil, err
		}
	} else {
		configs.CreatureTypes = make(map[string]*CreatureTypeConfig)
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
