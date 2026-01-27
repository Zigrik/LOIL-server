package game

import (
	"LOIL-server/internal/network"
)

// NetworkGame расширяет Game сетевыми возможностями
type NetworkGame struct {
	*Game
	Server *network.Server
}

// NewNetworkGame создает новую сетевую игру
func NewNetworkGame(w *World) *NetworkGame {
	game := NewGame(w)

	return &NetworkGame{
		Game:   game,
		Server: network.NewServer(game),
	}
}

// RunWithNetwork запускает игру с сетевым сервером
func (ng *NetworkGame) RunWithNetwork(addr string) error {
	// Запускаем игровой цикл в горутине
	go ng.RunGameLoop()

	// Запускаем WebSocket сервер
	return ng.Server.Run(addr)
}

// GetCharacterView возвращает представление персонажа для сети
func (g *Game) GetCharacterView(char *Character) *network.CharacterView {
	return &network.CharacterView{
		ID:         char.ID,
		Name:       char.Name,
		X:          char.X,
		Direction:  char.Direction,
		Speed:      char.Speed,
		Controlled: char.Controlled,
	}
}

// GetCreatureView возвращает представление существа для сети
func (g *Game) GetCreatureView(creature *Creature) *network.CreatureView {
	behaviorType := ""
	if creature.CurrentBehavior != nil {
		behaviorType = creature.CurrentBehavior.Type
	}

	return &network.CreatureView{
		ID:        creature.ID,
		TypeID:    creature.TypeID,
		Name:      creature.Name,
		X:         creature.X,
		Health:    creature.Health,
		MaxHealth: creature.MaxHealth,
		Behavior:  behaviorType,
	}
}

// GetObjectView возвращает представление объекта для сети
func (g *Game) GetObjectView(obj *WorldObject) *network.ObjectView {
	objConfig := g.GetObjectConfig(obj.TypeID)
	maxDurability := 0
	if objConfig != nil {
		maxDurability = objConfig.MaxDurability
	}

	return &network.ObjectView{
		ID:            obj.ID,
		TypeID:        obj.TypeID,
		X:             obj.X,
		Durability:    obj.Durability,
		MaxDurability: maxDurability,
	}
}

// GetLocationView возвращает представление локации для сети
func (g *Game) GetLocationView(locID int) *network.LocationView {
	locState := g.State.LocationStates[locID]
	if locState == nil {
		return nil
	}

	loc := g.GetLocation(locID)
	if loc == nil {
		return nil
	}

	return &network.LocationView{
		ID:         loc.ID,
		Name:       loc.Name,
		Foreground: locState.Foreground,
		Road:       locState.Road,
		Ground:     locState.Ground,
		Background: locState.Background,
		Width:      len(locState.Road),
	}
}
