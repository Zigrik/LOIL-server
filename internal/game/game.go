package game

import (
	"LOIL-server/internal/config"
	"LOIL-server/internal/world"
	"fmt"
	"time"
)

type Game struct {
	World      *world.World
	State      *GameState
	Registries *world.Registries
	ExitChan   chan bool
	UpdateChan chan bool
	InputChan  chan string
}

func NewGame(w *world.World) *Game {
	// Создаем реестры из конфигов
	registries := world.NewRegistries(w.Configs)

	state := &GameState{
		LocationStates:    make(map[int]*LocationState),
		CharsByLocation:   make(map[int][]*world.Character),
		ObjectsByLocation: make(map[int][]*world.WorldObject),
		Running:           true,
	}

	return &Game{
		World:      w,
		State:      state,
		Registries: registries,
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
		g.State.ObjectsByLocation[loc.ID] = []*world.WorldObject{}
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

	// Распределяем объекты по локациям
	for _, obj := range g.World.Objects {
		if obj.LocationID > 0 {
			g.State.ObjectsByLocation[obj.LocationID] = append(g.State.ObjectsByLocation[obj.LocationID], obj)
		}
	}
}

// GetObjectAtPosition возвращает объект на позиции
func (g *Game) GetObjectAtPosition(locationID int, pos int) *world.WorldObject {
	for _, obj := range g.State.ObjectsByLocation[locationID] {
		if obj.X == pos {
			return obj
		}
	}
	return nil
}

// GetObjectConfig возвращает конфигурацию объекта
func (g *Game) GetObjectConfig(typeID int) *config.ObjectTypeConfig {
	return g.Registries.GetObjectTypeConfig(typeID)
}

// GetRoadConfig возвращает конфигурацию дороги
func (g *Game) GetRoadConfig(typeID int) *config.RoadTypeConfig {
	return g.Registries.GetRoadTypeConfig(typeID)
}

// GetGroundConfig возвращает конфигурацию земли
func (g *Game) GetGroundConfig(typeID int) *config.GroundTypeConfig {
	return g.Registries.GetGroundTypeConfig(typeID)
}

// CheckRoadMovement проверяет возможность движения
func (g *Game) CheckRoadMovement(char *world.Character, pos int) (bool, float64) {
	locState := g.State.LocationStates[char.Location]
	if pos < 0 || pos >= len(locState.Road) {
		return false, 0.0
	}

	roadID := locState.Road[pos]

	// Если дорога -1 - движение невозможно
	if roadID == -1 {
		return false, 0.0
	}

	roadConfig := g.GetRoadConfig(roadID)
	if roadConfig == nil {
		return true, 0.7 // Базовая скорость
	}

	// Проверяем тип земли
	groundID := locState.Ground[pos]
	groundConfig := g.GetGroundConfig(groundID)
	if groundConfig != nil && !groundConfig.Walkable {
		return false, 0.0
	}

	return true, roadConfig.SpeedMod
}

// GetInteractions возвращает доступные взаимодействия на позиции
func (g *Game) GetInteractions(char *world.Character, pos int) []config.Interaction {
	var interactions []config.Interaction

	// Проверяем объект на позиции
	obj := g.GetObjectAtPosition(char.Location, pos)
	if obj != nil {
		objConfig := g.GetObjectConfig(obj.TypeID)
		if objConfig != nil {
			interactions = append(interactions, objConfig.Interactions...)
		}
	}

	return interactions
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
			// Проверка возможности движения по дороге
			canMove, speedMod := g.CheckRoadMovement(char, newPos)
			if !canMove {
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

			// Применяем модификатор скорости дороги
			char.Speed = 0.7 * speedMod

			// Проверяем доступные взаимодействия на новой позиции
			interactions := g.GetInteractions(char, newPos)
			if len(interactions) > 0 {
				// Получаем объект на позиции для вывода информации
				obj := g.GetObjectAtPosition(char.Location, newPos)
				if obj != nil {
					objConfig := g.GetObjectConfig(obj.TypeID)
					if objConfig != nil {
						fmt.Printf("%s находится возле %s\n", char.Name, objConfig.Name)
					}
				}
			}

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

func (g *Game) PrintState() {
	fmt.Println("\n=== СОСТОЯНИЕ МИРА ===")
	fmt.Printf("ID игрока: %d\n", g.World.PlayerID)

	for _, loc := range g.World.Locations {
		fmt.Printf("\nЛокация %d: %s\n", loc.ID, loc.Name)
		locState := g.State.LocationStates[loc.ID]

		// Выводим дорожный слой с персонажами
		fmt.Println("Дорожный слой с персонажами:")
		fmt.Print("Дорога: [")
		for i := 0; i < len(locState.Road); i++ {
			if locState.Foreground[i] != 0 {
				// Проверяем, персонаж ли это
				isCharacter := false
				for _, char := range g.State.CharsByLocation[loc.ID] {
					if char.ID == locState.Foreground[i] && int(char.X+0.5) == i {
						fmt.Printf("%c", char.Name[0])
						isCharacter = true
						break
					}
				}
				if !isCharacter {
					// Это объект на переднем плане
					objConfig := g.GetObjectConfig(locState.Foreground[i])
					if objConfig != nil {
						fmt.Print(objConfig.Name[0:1])
					} else {
						fmt.Print("?")
					}
				}
			} else if locState.Road[i] == -1 {
				fmt.Print("#") // Нет дороги
			} else {
				roadConfig := g.GetRoadConfig(locState.Road[i])
				if roadConfig != nil {
					fmt.Print(roadConfig.Name[0:1]) // Первая буква названия
				} else {
					fmt.Print("?")
				}
			}
			if i < len(locState.Road)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Println("]")

		// Выводим слой земли
		fmt.Print("Земля:  [")
		for i := 0; i < len(locState.Ground); i++ {
			groundConfig := g.GetGroundConfig(locState.Ground[i])
			if groundConfig != nil {
				fmt.Print(groundConfig.Name[0:1]) // Первая буква названия
			} else {
				fmt.Print("?")
			}
			if i < len(locState.Ground)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Println("]")

		// Выводим задний фон
		fmt.Print("Задний фон: [")
		for i := 0; i < len(locState.Background); i++ {
			if locState.Background[i] == 0 {
				fmt.Print(" ")
			} else {
				objConfig := g.GetObjectConfig(locState.Background[i])
				if objConfig != nil {
					fmt.Print(objConfig.Name[0:1]) // Первая буква названия
				} else {
					fmt.Print("?")
				}
			}
			if i < len(locState.Background)-1 {
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
				fmt.Printf("  %s (ID: %d) поз: %.1f, напр: %d, верт: %d, скорость: %.1f [%s]\n",
					char.Name, char.ID, char.X, char.Direction, char.Vertical, char.Speed, controlStatus)
			}
		}

		// Объекты в этой локации
		if objects, ok := g.State.ObjectsByLocation[loc.ID]; ok && len(objects) > 0 {
			fmt.Println("Объекты:")
			for _, obj := range objects {
				objConfig := g.GetObjectConfig(obj.TypeID)
				if objConfig != nil {
					fmt.Printf("  %s (ID: %d) тип: %d, поз: %d, прочность: %d/%d\n",
						objConfig.Name, obj.ID, obj.TypeID, obj.X, obj.Durability, objConfig.MaxDurability)
				} else {
					fmt.Printf("  Объект (ID: %d) тип: %d, поз: %d\n", obj.ID, obj.TypeID, obj.X)
				}
			}
		}
	}

	// Выводим информацию о реестрах
	if g.Registries != nil {
		fmt.Println("\n=== ИНФОРМАЦИЯ О РЕЕСТРАХ ===")
		fmt.Printf("Типов объектов: %d, Типов дорог: %d, Типов земли: %d\n",
			len(g.Registries.ObjectTypeByID),
			len(g.Registries.RoadTypeByID),
			len(g.Registries.GroundTypeByID))
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
		// Сохраняем только игровое состояние (без конфигов)
		saveWorld := &world.World{
			PlayerID:   g.World.PlayerID,
			Characters: g.World.Characters,
			Locations:  g.World.Locations,
			Objects:    g.World.Objects,
		}
		if err := world.SaveWorld(saveWorld, "data/save/world.json"); err != nil {
			fmt.Printf("Ошибка сохранения: %v\n", err)
		} else {
			fmt.Println("Мир сохранен в data/save/world.json")
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

// Helper function for layer printing
func (g *Game) PrintLayer(name string, layer []int, getSymbol func(int) string) {
	fmt.Printf("%s: [", name)
	for i, id := range layer {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(getSymbol(id))
	}
	fmt.Println("]")
}
