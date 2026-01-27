// internal/game/network_bridge.go
package game

import (
	"LOIL-server/internal/network"
	"time"
)

// GameNetworkBridge реализует интерфейс network.GameStateProvider
type GameNetworkBridge struct {
	Game *Game
}

// NewGameNetworkBridge создает мост между игрой и сетью
func NewGameNetworkBridge(game *Game) *GameNetworkBridge {
	return &GameNetworkBridge{Game: game}
}

// GetLocationState возвращает состояние локации для сети
func (b *GameNetworkBridge) GetLocationState(locationID int) *network.LocationState {
	locState := b.Game.State.LocationStates[locationID]
	if locState == nil {
		return nil
	}

	loc := b.Game.GetLocation(locationID)
	if loc == nil {
		return nil
	}

	return &network.LocationState{
		ID:         loc.ID,
		Name:       loc.Name,
		Width:      len(locState.Road),
		Foreground: locState.Foreground,
		Road:       locState.Road,
		Ground:     locState.Ground,
		Background: locState.Background,
		LastUpdate: time.Now().UnixMilli(),
	}
}

// GetCharactersInLocation возвращает персонажей в локации
func (b *GameNetworkBridge) GetCharactersInLocation(locationID int) []*network.CharacterState {
	chars, ok := b.Game.State.CharsByLocation[locationID]
	if !ok {
		return nil
	}

	result := make([]*network.CharacterState, 0, len(chars))
	for _, char := range chars {
		result = append(result, b.characterToNetwork(char))
	}

	return result
}

// GetCreaturesInLocation возвращает существ в локации
func (b *GameNetworkBridge) GetCreaturesInLocation(locationID int) []*network.CreatureState {
	creatures, ok := b.Game.State.CreaturesByLocation[locationID]
	if !ok {
		return nil
	}

	result := make([]*network.CreatureState, 0, len(creatures))
	for _, creature := range creatures {
		result = append(result, b.creatureToNetwork(creature))
	}

	return result
}

// GetObjectsInLocation возвращает объекты в локации
func (b *GameNetworkBridge) GetObjectsInLocation(locationID int) []*network.ObjectState {
	objects, ok := b.Game.State.ObjectsByLocation[locationID]
	if !ok {
		return nil
	}

	result := make([]*network.ObjectState, 0, len(objects))
	for _, obj := range objects {
		result = append(result, b.objectToNetwork(obj))
	}

	return result
}

// GetCharacterByID возвращает персонажа по ID
func (b *GameNetworkBridge) GetCharacterByID(characterID int) *network.CharacterState {
	char := b.Game.GetPlayerCharacter() // TODO: Нужен метод получения по ID
	if char != nil && char.ID == characterID {
		return b.characterToNetwork(char)
	}
	return nil
}

// HandleJoin обрабатывает присоединение игрока
func (b *GameNetworkBridge) HandleJoin(playerID, characterID, locationID int) (*network.CharacterState, error) {
	// TODO: Реализовать логику присоединения
	// Пока просто возвращаем первого персонажа
	char := b.Game.GetPlayerCharacter()
	if char == nil {
		return nil, network.NewError("no_character", "Персонаж не найден")
	}

	return b.characterToNetwork(char), nil
}

// HandleMove обрабатывает движение
func (b *GameNetworkBridge) HandleMove(playerID int, direction, vertical int) error {
	char := b.Game.GetPlayerCharacter()
	if char == nil {
		return network.NewError("no_character", "Персонаж не найден")
	}

	// Преобразуем команду в формат игры
	if direction == 0 {
		char.Direction = 0
		char.Vertical = 0
	} else {
		char.Direction = direction
		char.Vertical = vertical
	}

	return nil
}

// HandleStop обрабатывает остановку
func (b *GameNetworkBridge) HandleStop(playerID int) error {
	char := b.Game.GetPlayerCharacter()
	if char == nil {
		return network.NewError("no_character", "Персонаж не найден")
	}

	char.Direction = 0
	char.Vertical = 0

	return nil
}

// HandleInteract обрабатывает взаимодействие
func (b *GameNetworkBridge) HandleInteract(playerID int, objectID, interactionIdx int) (*network.InteractionResult, error) {
	char := b.Game.GetPlayerCharacter()
	if char == nil {
		return nil, network.NewError("no_character", "Персонаж не найден")
	}

	// TODO: Реализовать взаимодействие через существующую логику
	// b.Game.PerformInteractionByIndex(char, objectID, interactionIdx)

	return &network.InteractionResult{
		Success:    true,
		ObjectID:   objectID,
		Message:    "Взаимодействие выполнено",
		ServerTime: time.Now().UnixMilli(),
	}, nil
}

// GetServerTime возвращает время сервера
func (b *GameNetworkBridge) GetServerTime() int64 {
	return time.Now().UnixMilli()
}

// GetLocationName возвращает название локации
func (b *GameNetworkBridge) GetLocationName(locationID int) string {
	loc := b.Game.GetLocation(locationID)
	if loc != nil {
		return loc.Name
	}
	return ""
}

// Вспомогательные методы преобразования
func (b *GameNetworkBridge) characterToNetwork(char *Character) *network.CharacterState {
	return &network.CharacterState{
		ID:         char.ID,
		Name:       char.Name,
		LocationID: char.Location,
		X:          char.X,
		Direction:  char.Direction,
		Speed:      char.Speed,
		Controlled: char.Controlled,
		LastUpdate: time.Now().UnixMilli(),
	}
}

func (b *GameNetworkBridge) creatureToNetwork(creature *Creature) *network.CreatureState {
	behavior := ""
	if creature.CurrentBehavior != nil {
		behavior = creature.CurrentBehavior.Type
	}

	return &network.CreatureState{
		ID:         creature.ID,
		TypeID:     creature.TypeID,
		Name:       creature.Name,
		LocationID: creature.Location,
		X:          creature.X,
		Health:     creature.Health,
		MaxHealth:  creature.MaxHealth,
		Hunger:     creature.Hunger,
		Behavior:   behavior,
		LastUpdate: creature.LastUpdate.UnixMilli(),
	}
}

func (b *GameNetworkBridge) objectToNetwork(obj *WorldObject) *network.ObjectState {
	objConfig := b.Game.GetObjectConfig(obj.TypeID)
	maxDurability := 0
	if objConfig != nil {
		maxDurability = objConfig.MaxDurability
	}

	return &network.ObjectState{
		ID:            obj.ID,
		TypeID:        obj.TypeID,
		LocationID:    obj.LocationID,
		X:             obj.X,
		Durability:    obj.Durability,
		MaxDurability: maxDurability,
		GrowthStage:   obj.GrowthStage,
		LastUpdate:    time.Now().UnixMilli(),
	}
}
