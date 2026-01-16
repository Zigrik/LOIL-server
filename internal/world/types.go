package world

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RoadTiles []int

// MarshalJSON преобразует RoadTiles в строку для JSON
func (rt RoadTiles) MarshalJSON() ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("\"")
	for i, tile := range rt {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%d", tile))
	}
	sb.WriteString("\"")
	return []byte(sb.String()), nil
}

// UnmarshalJSON преобразует строку из JSON в RoadTiles
func (rt *RoadTiles) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	// Разбиваем строку по пробелам
	parts := strings.Fields(str)
	*rt = make(RoadTiles, len(parts))

	for i, part := range parts {
		var tile int
		if _, err := fmt.Sscanf(part, "%d", &tile); err != nil {
			return fmt.Errorf("ошибка парсинга тайла '%s': %v", part, err)
		}
		(*rt)[i] = tile
	}

	return nil
}

type Transition struct {
	LocationID int    `json:"location_id"`
	Type       string `json:"type"`
}

// Общий тип для всех слоев
type IntSlice []int

// MarshalJSON для всех слоев
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

// UnmarshalJSON для всех слоев
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

type Location struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Foreground  IntSlice               `json:"foreground"` // Передний фон (объекты, персонажи)
	Road        IntSlice               `json:"road"`       // Дорожный слой (-1 = нет дороги, 0+ = разрешено)
	Ground      IntSlice               `json:"ground"`     // Слой земли (0=земля, 1=песок, 2=глина, 3=камень, -1=ручей, -2=река)
	Background  IntSlice               `json:"background"` // Задний фон (деревья, дома)
	Transitions map[string]*Transition `json:"transitions"`
}

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

type World struct {
	PlayerID   int          `json:"player_id"`
	Characters []*Character `json:"characters"`
	Locations  []*Location  `json:"locations"`
}
