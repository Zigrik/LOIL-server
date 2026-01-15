package game

import (
	"LOIL-server/internal/world"
	"fmt"
	"time"
)

type GameState struct {
	LocationTiles   map[int][]int
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

func NewGame(w *world.World) *Game {
	state := &GameState{
		LocationTiles:   make(map[int][]int),
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
	// Инициализируем тайлы локаций
	for _, loc := range g.World.Locations {
		g.State.LocationTiles[loc.ID] = make([]int, len(loc.RoadTiles))
		copy(g.State.LocationTiles[loc.ID], loc.RoadTiles)
		g.State.CharsByLocation[loc.ID] = []*world.Character{}
	}

	// Распределяем персонажей по локациям
	for _, char := range g.World.Characters {
		pos := int(char.X + 0.5)
		if pos >= 0 && pos < len(g.State.LocationTiles[char.Location]) {
			g.State.LocationTiles[char.Location][pos] = char.ID
			g.State.CharsByLocation[char.Location] = append(g.State.CharsByLocation[char.Location], char)
		}
	}
}

func (g *Game) UpdateCharacter(char *world.Character, elapsed float64) bool {
	if !g.State.Running {
		return false
	}

	locID := char.Location
	tiles := g.State.LocationTiles[locID]

	if len(tiles) == 0 {
		return false
	}

	oldPos := int(char.X + 0.5)

	if char.Direction != 0 {
		char.X += float64(char.Direction) * char.Speed * elapsed

		// Проверка границ локации
		if char.Direction == 1 && char.X >= float64(len(tiles)-1) {
			char.X = float64(len(tiles) - 1)
			g.TryTransition(char, "right")
			return true
		} else if char.Direction == -1 && char.X <= 0 {
			char.X = 0
			g.TryTransition(char, "left")
			return true
		}

		// Обновление позиции в тайлах
		newPos := int(char.X + 0.5)
		if newPos != oldPos && newPos >= 0 && newPos < len(tiles) {
			// Проверка столкновения
			if tiles[newPos] != 0 && tiles[newPos] != char.ID {
				fmt.Printf("\n[СТОЛКНОВЕНИЕ] На клетке %d персонаж с ID: %d\n", newPos, tiles[newPos])
				char.X = float64(oldPos)
				return false
			}

			// Освобождаем старую позицию
			if oldPos >= 0 && oldPos < len(tiles) {
				tiles[oldPos] = 0
			}

			// Занимаем новую позицию
			tiles[newPos] = char.ID
			return true
		}
	}
	return false
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
			// Если нет вертикального направления, остаемся на месте
			char.Direction = 0
			return
		}
	} else if side == "right" {
		if char.Vertical == 1 {
			transitionKey = "right_up"
		} else if char.Vertical == -1 {
			transitionKey = "right_down"
		} else {
			// Если нет вертикального направления, остаемся на месте
			char.Direction = 0
			return
		}
	}

	if trans, ok := loc.Transitions[transitionKey]; ok && trans != nil {
		// Удаляем из старой локации
		oldLocID := char.Location
		oldPos := int(char.X + 0.5)
		if oldPos >= 0 && oldPos < len(g.State.LocationTiles[oldLocID]) {
			g.State.LocationTiles[oldLocID][oldPos] = 0
		}
		g.State.CharsByLocation[oldLocID] = g.removeCharFromSlice(g.State.CharsByLocation[oldLocID], char)

		// Перемещаем в новую локацию
		char.Location = trans.LocationID
		newTiles := g.State.LocationTiles[char.Location]

		if side == "left" {
			char.X = float64(len(newTiles) - 1)
			char.Direction = 0 // Останавливаемся после перехода
		} else {
			char.X = 0
			char.Direction = 0 // Останавливаемся после перехода
		}
		char.Vertical = 0 // Сбрасываем вертикальное направление

		// Добавляем в новую локацию
		newPos := int(char.X + 0.5)
		if newPos >= 0 && newPos < len(newTiles) {
			newTiles[newPos] = char.ID
		}
		g.State.CharsByLocation[char.Location] = append(g.State.CharsByLocation[char.Location], char)

		fmt.Printf("\n%s перешел в локацию %d\n", char.Name, char.Location)
		g.UpdateChan <- true
	} else {
		// Если перехода нет, останавливаемся
		char.Direction = 0
		fmt.Printf("\n%s достиг края, но перехода нет\n", char.Name)
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

func (g *Game) PrintState() {
	fmt.Println("\n=== СОСТОЯНИЕ МИРА ===")
	fmt.Printf("ID игрока: %d\n", g.World.PlayerID)

	for _, loc := range g.World.Locations {
		fmt.Printf("\nЛокация %d: %s\n", loc.ID, loc.Name)
		tiles := g.State.LocationTiles[loc.ID]

		fmt.Print("[")
		for i := 0; i < len(tiles); i++ {
			if tiles[i] == 0 {
				fmt.Print(".")
			} else {
				for _, char := range g.State.CharsByLocation[loc.ID] {
					if char.ID == tiles[i] && int(char.X+0.5) == i {
						fmt.Printf("%c", char.Name[0])
						break
					}
				}
			}
			if i < len(tiles)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Println("]")

		// Персонажи в этой локации
		if chars, ok := g.State.CharsByLocation[loc.ID]; ok && len(chars) > 0 {
			fmt.Println("Персонажи:")
			for _, char := range chars {
				controlStatus := "NPC"
				if char.Controlled == g.World.PlayerID {
					controlStatus = "ИГРОК"
				}
				fmt.Printf("  %s (ID: %d) поз: %.1f, напр: %d, верт: %d [%s]\n",
					char.Name, char.ID, char.X, char.Direction, char.Vertical, controlStatus)
			}
		}
	}
	fmt.Println("\nКоманды: a/d - влево/вправо, w/s - вверх/вниз, stop - остановка, x - состояние, save - сохранить, exit - выход")
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
	case "x":
		g.PrintState()
	case "save":
		if err := world.SaveWorld(g.World, "world.json"); err != nil {
			fmt.Printf("Ошибка сохранения: %v\n", err)
		} else {
			fmt.Println("Мир сохранен в world.json")
		}
	case "exit":
		g.State.Running = false
		g.ExitChan <- true
	default:
		fmt.Println("Неизвестная команда. Доступные: a, d, w, s, stop, x, save, exit")
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
