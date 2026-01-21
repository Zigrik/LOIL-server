package world

import "LOIL-server/internal/config"

// Registries - реестры для быстрого доступа
type Registries struct {
	ObjectTypeByID   map[int]*config.ObjectTypeConfig
	RoadTypeByID     map[int]*config.RoadTypeConfig
	GroundTypeByID   map[int]*config.GroundTypeConfig
	ItemTypeByID     map[int]*config.ItemTypeConfig
	CreatureTypeByID map[int]*config.CreatureTypeConfig
}

// NewRegistries создает реестры из конфигов
func NewRegistries(configs *config.Configs) *Registries {
	r := &Registries{
		ObjectTypeByID:   make(map[int]*config.ObjectTypeConfig),
		RoadTypeByID:     make(map[int]*config.RoadTypeConfig),
		GroundTypeByID:   make(map[int]*config.GroundTypeConfig),
		ItemTypeByID:     make(map[int]*config.ItemTypeConfig),
		CreatureTypeByID: make(map[int]*config.CreatureTypeConfig),
	}

	// Заполняем реестры объектов
	for _, objType := range configs.ObjectTypes {
		r.ObjectTypeByID[objType.ID] = objType
	}

	// Заполняем реестры дорог
	for _, roadType := range configs.RoadTypes {
		r.RoadTypeByID[roadType.ID] = roadType
	}

	// Заполняем реестры земли
	for _, groundType := range configs.GroundTypes {
		r.GroundTypeByID[groundType.ID] = groundType
	}

	// Заполняем реестры предметов
	for _, itemType := range configs.ItemTypes {
		r.ItemTypeByID[itemType.ID] = itemType
	}

	// Заполняем реестры существ
	for _, creatureType := range configs.CreatureTypes {
		r.CreatureTypeByID[creatureType.ID] = creatureType
	}

	return r
}

// GetObjectTypeConfig возвращает конфиг типа объекта
func (r *Registries) GetObjectTypeConfig(typeID int) *config.ObjectTypeConfig {
	return r.ObjectTypeByID[typeID]
}

// GetRoadTypeConfig возвращает конфиг типа дороги
func (r *Registries) GetRoadTypeConfig(typeID int) *config.RoadTypeConfig {
	return r.RoadTypeByID[typeID]
}

// GetGroundTypeConfig возвращает конфиг типа земли
func (r *Registries) GetGroundTypeConfig(typeID int) *config.GroundTypeConfig {
	return r.GroundTypeByID[typeID]
}

// GetItemTypeConfig возвращает конфиг типа предмета
func (r *Registries) GetItemTypeConfig(typeID int) *config.ItemTypeConfig {
	return r.ItemTypeByID[typeID]
}

// GetCreatureTypeConfig возвращает конфиг типа существа
func (r *Registries) GetCreatureTypeConfig(typeID int) *config.CreatureTypeConfig {
	return r.CreatureTypeByID[typeID]
}
