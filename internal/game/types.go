package game

import "LOIL-server/internal/world"

// LocationState - состояние локации в игре
type LocationState struct {
	Foreground []int // Положительные значения: объекты, персонажи; Отрицательные: существа
	Road       []int
	Ground     []int
	Background []int
}

// GameState - состояние игры
type GameState struct {
	LocationStates      map[int]*LocationState       // Состояние слоев каждой локации
	CharsByLocation     map[int][]*world.Character   // Персонажи по локациям
	ObjectsByLocation   map[int][]*world.WorldObject // Объекты по локациям
	CreaturesByLocation map[int][]*world.Creature    // Существа по локациям
	LastUpdate          int64                        // Время последнего обновления
	Running             bool                         // Запущена ли игра
}

// CreatureBehaviorInfo - информация о поведении существа для отображения
type CreatureBehaviorInfo struct {
	CreatureID   int
	CreatureName string
	BehaviorType string
	TargetPos    int
	Progress     float64 // Прогресс выполнения поведения (0-1)
}
