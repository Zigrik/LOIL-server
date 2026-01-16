package game

import (
	"LOIL-server/internal/world"
	"fmt"
	"time"
)

// Структура для хранения состояния всех слоев локации
type LocationState struct {
	Foreground []int
	Road       []int
	Ground     []int
	Background []int
}

type GameState struct {
	LocationStates  map[int]*LocationState
	CharsByLocation map[int][]*world.Character
	LastUpdate      int64
	Running         bool
}

type Game struct {
	World      *world.World
	State      *GameState
	ExitChan   chan bool
	UpdateChan chan bool
	InputChan  chan string
}

// GetDebugLocationStates возвращает состояния локаций для отладки
func (g *Game) GetDebugLocationStates() map[int]*DebugLocationState {
	result := make(map[int]*DebugLocationState)

	for id, state := range g.State.LocationStates {
		result[id] = &DebugLocationState{
			Foreground: state.Foreground,
			Road:       state.Road,
			Ground:     state.Ground,
			Background: state.Background,
		}
	}

	return result
}

// DebugLocationState - структура для отладочного вывода
type DebugLocationState struct {
	Foreground []int
	Road       []int
	Ground     []int
	Background []int
}

func NewGame(w *world.World) *Game {
	state := &GameState{
		LocationStates:  make(map[int]*LocationState),
		CharsByLocation: make(map[int][]*world.Character),
		Running:         true,
	}

	return &Game{
		World:      w,
		State:      state,
		ExitChan:   make(chan bool),
		UpdateChan: make(chan bool, 100),
		InputChan:  make(chan string, 10),
	}
}

func (g *Game) Initialize() {
	// Инициализируем все слои для каждой локации
	for _, loc := range g.World.Locations {
		state := &LocationState{
			Foreground: make([]int, len(loc.Foreground)),
			Road:       make([]int, len(loc.Road)),
			Ground:     make([]int, len(loc.Ground)),
			Background: make([]int, len(loc.Background)),
		}

		// Копируем данные из мира
		copy(state.Foreground, []int(loc.Foreground))
		copy(state.Road, []int(loc.Road))
		copy(state.Ground, []int(loc.Ground))
		copy(state.Background, []int(loc.Background))

		g.State.LocationStates[loc.ID] = state
		g.State.CharsByLocation[loc.ID] = []*world.Character{}
	}

	// Распределяем персонажей по локациям
	for _, char := range g.World.Characters {
		pos := int(char.X + 0.5)
		locState := g.State.LocationStates[char.Location]

		if pos >= 0 && pos < len(locState.Road) {
			// Сохраняем ID персонажа в слое переднего плана
			locState.Foreground[pos] = char.ID
			g.State.CharsByLocation[char.Location] = append(g.State.CharsByLocation[char.Location], char)
		}
	}
}

func (g *Game) UpdateCharacter(char *world.Character, elapsed float64) bool {
	if !g.State.Running {
		return false
	}

	locID := char.Location
	locState := g.State.LocationStates[locID]
	roadLayer := locState.Road

	if len(roadLayer) == 0 {
		return false
	}

	oldPos := int(char.X + 0.5)

	if char.Direction != 0 {
		char.X += float64(char.Direction) * char.Speed * elapsed

		// Проверка границ локации
		if char.Direction == 1 && char.X >= float64(len(roadLayer)-1) {
			char.X = float64(len(roadLayer) - 1)
			g.TryTransition(char, "right")
			return true
		} else if char.Direction == -1 && char.X <= 0 {
			char.X = 0
			g.TryTransition(char, "left")
			return true
		}

		// Обновление позиции в тайлах
		newPos := int(char.X + 0.5)
		if newPos != oldPos && newPos >= 0 && newPos < len(roadLayer) {
			// Проверка дороги
			if roadLayer[newPos] == -1 {
				// Нет дороги - движение невозможно
				fmt.Printf("[ПРЕПЯТСТВИЕ] %s не может идти по этой местности\n", char.Name)
				char.X = float64(oldPos)
				return false
			}

			// Проверка столкновения с другими персонажами
			if locState.Foreground[newPos] != 0 && locState.Foreground[newPos] != char.ID {
				fmt.Printf("[СТОЛКНОВЕНИЕ] На клетке %d персонаж с ID: %d\n", newPos, locState.Foreground[newPos])
				char.X = float64(oldPos)
				return false
			}

			// Освобождаем старую позицию
			if oldPos >= 0 && oldPos < len(locState.Foreground) {
				locState.Foreground[oldPos] = 0
			}

			// Занимаем новую позицию
			locState.Foreground[newPos] = char.ID

			// Проверяем тип поверхности под ногами
			g.CheckGroundType(char, newPos)

			return true
		}
	}
	return false
}

func (g *Game) CheckGroundType(char *world.Character, pos int) {
	locState := g.State.LocationStates[char.Location]
	if pos < 0 || pos >= len(locState.Ground) {
		return
	}

	groundType := locState.Ground[pos]

	switch groundType {
	case -2: // Река
		char.Speed = 0.3 // Замедление в реке
	case -1: // Ручей
		char.Speed = 0.5 // Среднее замедление
	case 1: // Песок
		char.Speed = 0.6 // Легкое замедление
	case 2: // Глина
		char.Speed = 0.8 // Почти нормально
	case 3: // Камень
		char.Speed = 1.0 // Нормальная скорость
	default: // Земля (0)
		char.Speed = 0.7 // Базовая скорость
	}
}

func (g *Game) TryTransition(char *world.Character, side string) {
	loc := g.GetLocation(char.Location)
	if loc == nil || loc.Transitions == nil {
		return
	}

	var transitionKey string
	if side == "left" {
		if char.Vertical == 1 {
			transitionKey = "left_up"
		} else if char.Vertical == -1 {
			transitionKey = "left_down"
		} else {
			char.Direction = 0
			return
		}
	} else if side == "right" {
		if char.Vertical == 1 {
			transitionKey = "right_up"
		} else if char.Vertical == -1 {
			transitionKey = "right_down"
		} else {
			char.Direction = 0
			return
		}
	}

	if trans, ok := loc.Transitions[transitionKey]; ok && trans != nil {
		// Удаляем из старой локации
		oldLocID := char.Location
		oldPos := int(char.X + 0.5)
		oldLocState := g.State.LocationStates[oldLocID]

		if oldPos >= 0 && oldPos < len(oldLocState.Foreground) {
			oldLocState.Foreground[oldPos] = 0
		}
		g.State.CharsByLocation[oldLocID] = g.removeCharFromSlice(g.State.CharsByLocation[oldLocID], char)

		// Перемещаем в новую локацию
		char.Location = trans.LocationID
		newLocState := g.State.LocationStates[char.Location]

		if side == "left" {
			char.X = float64(len(newLocState.Road) - 1)
			char.Direction = 0
		} else {
			char.X = 0
			char.Direction = 0
		}
		char.Vertical = 0
		char.Speed = 0.7 // Сбрасываем скорость к базовой

		// Добавляем в новую локацию
		newPos := int(char.X + 0.5)
		if newPos >= 0 && newPos < len(newLocState.Foreground) {
			newLocState.Foreground[newPos] = char.ID
		}
		g.State.CharsByLocation[char.Location] = append(g.State.CharsByLocation[char.Location], char)

		fmt.Printf("%s перешел в локацию %d\n", char.Name, char.Location)
		g.UpdateChan <- true
	} else {
		char.Direction = 0
		fmt.Printf("%s достиг края, но перехода нет\n", char.Name)
	}
}

func (g *Game) removeCharFromSlice(slice []*world.Character, char *world.Character) []*world.Character {
	for i, c := range slice {
		if c.ID == char.ID {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func (g *Game) GetLocation(id int) *world.Location {
	for _, loc := range g.World.Locations {
		if loc.ID == id {
			return loc
		}
	}
	return nil
}

func (g *Game) GetPlayerCharacter() *world.Character {
	for _, char := range g.World.Characters {
		if char.Controlled == g.World.PlayerID {
			return char
		}
	}
	return nil
}

func (g *Game) HandleInput(input string) {
	if !g.State.Running {
		return
	}

	playerChar := g.GetPlayerCharacter()
	if playerChar == nil {
		fmt.Println("У игрока нет управляемого персонажа")
		return
	}

	switch input {
	case "a":
		playerChar.Direction = -1
		playerChar.Vertical = 0 // Сбрасываем вертикаль при горизонтальном движении
		fmt.Printf("%s начал движение влево\n", playerChar.Name)
	case "d":
		playerChar.Direction = 1
		playerChar.Vertical = 0 // Сбрасываем вертикаль при горизонтальном движении
		fmt.Printf("%s начал движение вправо\n", playerChar.Name)
	case "w":
		playerChar.Vertical = 1
		fmt.Printf("%s готов к переходу вверх\n", playerChar.Name)
	case "s":
		playerChar.Vertical = -1
		fmt.Printf("%s готов к переходу вниз\n", playerChar.Name)
	case "stop":
		playerChar.Direction = 0
		playerChar.Vertical = 0
		fmt.Printf("%s остановился\n", playerChar.Name)
	case "exit":
		g.State.Running = false
		g.ExitChan <- true
	default:
		fmt.Println("Неизвестная команда. Доступные: a, d, w, s, stop, exit")
	}
}

func (g *Game) RunGameLoop() {
	lastUpdate := time.Now()

	for g.State.Running {
		select {
		case <-g.ExitChan:
			fmt.Println("Завершение игрового цикла...")
			return
		case input := <-g.InputChan:
			g.HandleInput(input)
		case <-time.After(16 * time.Millisecond):
			currentTime := time.Now()
			elapsed := currentTime.Sub(lastUpdate).Seconds()
			lastUpdate = currentTime

			updated := false
			for _, char := range g.World.Characters {
				if g.UpdateCharacter(char, elapsed) {
					updated = true
				}
			}

			if updated {
				select {
				case g.UpdateChan <- true:
				default:
				}
			}
		}
	}
}

// Методы для отладки (используются debug пакетом)

// GetLocationStates возвращает все состояния локаций
func (g *Game) GetLocationStates() map[int]*LocationState {
	return g.State.LocationStates
}

// GetLocationState возвращает состояние конкретной локации
func (g *Game) GetLocationState(locationID int) *LocationState {
	if state, ok := g.State.LocationStates[locationID]; ok {
		return state
	}
	return nil
}

// GetCharsByLocation возвращает персонажей по локациям
func (g *Game) GetCharsByLocation() map[int][]*world.Character {
	return g.State.CharsByLocation
}

// GetWorld возвращает мир
func (g *Game) GetWorld() *world.World {
	return g.World
}
