package game

import "LOIL-server/internal/world"

// LocationState - состояние локации в игре
type LocationState struct {
	Foreground []int
	Road       []int
	Ground     []int
	Background []int
}

// GameState - состояние игры
type GameState struct {
	LocationStates    map[int]*LocationState
	CharsByLocation   map[int][]*world.Character
	ObjectsByLocation map[int][]*world.WorldObject
	LastUpdate        int64
	Running           bool
}
