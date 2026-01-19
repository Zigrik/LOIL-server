package world

import "LOIL-server/internal/config"

// Registries - реестры для быстрого доступа
type Registries struct {
	ObjectTypeByID map[int]*config.ObjectTypeConfig
	RoadTypeByID   map[int]*config.RoadTypeConfig
	GroundTypeByID map[int]*config.GroundTypeConfig
}

// NewRegistries создает реестры из конфигов
func NewRegistries(configs *config.Configs) *Registries {
	r := &Registries{
		ObjectTypeByID: make(map[int]*config.ObjectTypeConfig),
		RoadTypeByID:   make(map[int]*config.RoadTypeConfig),
		GroundTypeByID: make(map[int]*config.GroundTypeConfig),
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
