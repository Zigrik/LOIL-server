package world

import (
	"LOIL-server/internal/config"
	"encoding/json"
	"os"
)

// LoadWorld загружает мир из файла
func LoadWorld(filename string, configs *config.Configs) (*World, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	world := &World{}
	if err := json.Unmarshal(data, world); err != nil {
		return nil, err
	}

	// Присваиваем конфиги
	world.Configs = configs

	// Инициализируем объекты если их нет
	if world.Objects == nil {
		world.Objects = make(map[int]*WorldObject)
	}

	// Собираем все объекты из локаций
	for _, loc := range world.Locations {
		if loc.Objects == nil {
			loc.Objects = make(map[int]*WorldObject)
		}

		// Добавляем объекты локации в общий список
		for objID, obj := range loc.Objects {
			world.Objects[objID] = obj
		}
	}

	return world, nil
}

// SaveWorld сохраняет мир в файл
func SaveWorld(world *World, filename string) error {
	data, err := json.MarshalIndent(world, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
