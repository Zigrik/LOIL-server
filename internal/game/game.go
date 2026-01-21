package game

import (
	"LOIL-server/internal/config"
	worldpkg "LOIL-server/internal/world"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Game struct {
	GameWorld  *worldpkg.World
	State      *GameState
	Registries *worldpkg.Registries
	ExitChan   chan bool
	UpdateChan chan bool
	InputChan  chan string
}

func NewGame(w *worldpkg.World) *Game {
	// Создаем реестры из конфигов
	registries := worldpkg.NewRegistries(w.Configs)

	state := &GameState{
		LocationStates:      make(map[int]*LocationState),
		CharsByLocation:     make(map[int][]*worldpkg.Character),
		ObjectsByLocation:   make(map[int][]*worldpkg.WorldObject),
		CreaturesByLocation: make(map[int][]*worldpkg.Creature),
		Running:             true,
	}

	return &Game{
		GameWorld:  w,
		State:      state,
		Registries: registries,
		ExitChan:   make(chan bool),
		UpdateChan: make(chan bool, 100),
		InputChan:  make(chan string, 10),
	}
}

func (g *Game) Initialize() {
	// Инициализируем все слои для каждой локации
	for _, loc := range g.GameWorld.Locations {
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
		g.State.CharsByLocation[loc.ID] = []*worldpkg.Character{}
		g.State.ObjectsByLocation[loc.ID] = []*worldpkg.WorldObject{}
		g.State.CreaturesByLocation[loc.ID] = []*worldpkg.Creature{}
	}

	// Распределяем персонажей по локациям
	for _, char := range g.GameWorld.Characters {
		pos := int(char.X + 0.5)
		locState := g.State.LocationStates[char.Location]

		if pos >= 0 && pos < len(locState.Road) {
			// Сохраняем ID персонажа в слое переднего плана
			locState.Foreground[pos] = char.ID
			g.State.CharsByLocation[char.Location] = append(g.State.CharsByLocation[char.Location], char)
		}
	}

	// Распределяем объекты по локациям
	for _, obj := range g.GameWorld.Objects {
		if obj.LocationID > 0 {
			g.State.ObjectsByLocation[obj.LocationID] = append(g.State.ObjectsByLocation[obj.LocationID], obj)
		}
	}

	// Распределяем существ по локациям
	for _, creature := range g.GameWorld.Creatures {
		pos := int(creature.X + 0.5)
		locState := g.State.LocationStates[creature.Location]

		if pos >= 0 && pos < len(locState.Road) {
			// Сохраняем ID существа (отрицательный для отличия от персонажей)
			locState.Foreground[pos] = -creature.ID
			g.State.CreaturesByLocation[creature.Location] = append(g.State.CreaturesByLocation[creature.Location], creature)
		}
	}

	// Инициализируем начальное поведение существ
	rand.Seed(time.Now().UnixNano())
	for _, creature := range g.GameWorld.Creatures {
		g.SetDefaultBehavior(creature)
		creature.LastUpdate = time.Now()
	}
}

// GetObjectAtPosition возвращает объект на позиции
func (g *Game) GetObjectAtPosition(locationID int, pos int) *worldpkg.WorldObject {
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

// GetItemConfig возвращает конфигурацию предмета
func (g *Game) GetItemConfig(typeID int) *config.ItemTypeConfig {
	return g.Registries.GetItemTypeConfig(typeID)
}

// GetCreatureConfig возвращает конфигурацию существа
func (g *Game) GetCreatureConfig(typeID int) *config.CreatureTypeConfig {
	return g.Registries.GetCreatureTypeConfig(typeID)
}

// GetCreatureAtPosition возвращает существо на позиции
func (g *Game) GetCreatureAtPosition(locationID int, pos int) *worldpkg.Creature {
	for _, creature := range g.State.CreaturesByLocation[locationID] {
		if int(creature.X+0.5) == pos {
			return creature
		}
	}
	return nil
}

// GetCreatureByID возвращает существо по ID
func (g *Game) GetCreatureByID(id int) *worldpkg.Creature {
	for _, creature := range g.GameWorld.Creatures {
		if creature.ID == id {
			return creature
		}
	}
	return nil
}

// CheckRoadMovement проверяет возможность движения
func (g *Game) CheckRoadMovement(char *worldpkg.Character, pos int) (bool, float64) {
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

	// Проверяем столкновение с другими персонажами или существами
	if locState.Foreground[pos] != 0 && locState.Foreground[pos] != char.ID {
		// Проверяем, существо ли это (отрицательный ID)
		if locState.Foreground[pos] < 0 {
			creatureID := -locState.Foreground[pos]
			if creature := g.GetCreatureByID(creatureID); creature != nil {
				creatureConfig := g.GetCreatureConfig(creature.TypeID)
				if creatureConfig != nil {
					fmt.Printf("[СТОЛКНОВЕНИЕ] На клетке %d %s (ID: %d)\n",
						pos, creatureConfig.Name, creature.ID)
				}
			}
		} else {
			fmt.Printf("[СТОЛКНОВЕНИЕ] На клетке %d персонаж с ID: %d\n",
				pos, locState.Foreground[pos])
		}
		return false, 0.0
	}

	return true, roadConfig.SpeedMod
}

// GetAvailableInteractions возвращает доступные взаимодействия для персонажа
func (g *Game) GetAvailableInteractions(char *worldpkg.Character) []config.Interaction {
	var interactions []config.Interaction

	// Получаем позицию персонажа
	pos := int(char.X + 0.5)

	// Проверяем объекты на текущей позиции
	if obj := g.GetObjectAtPosition(char.Location, pos); obj != nil {
		objConfig := g.GetObjectConfig(obj.TypeID)
		if objConfig != nil {
			// Фильтруем взаимодействия по доступному инструменту
			for _, interaction := range objConfig.Interactions {
				if g.CanPerformInteraction(char, interaction) {
					interactions = append(interactions, interaction)
				}
			}
		}
	}

	// Проверяем объекты на соседних клетках
	neighborPositions := []int{pos - 1, pos + 1}
	for _, neighborPos := range neighborPositions {
		if neighborPos >= 0 && neighborPos < len(g.State.LocationStates[char.Location].Road) {
			if obj := g.GetObjectAtPosition(char.Location, neighborPos); obj != nil {
				objConfig := g.GetObjectConfig(obj.TypeID)
				if objConfig != nil {
					for _, interaction := range objConfig.Interactions {
						if g.CanPerformInteraction(char, interaction) {
							interactions = append(interactions, interaction)
						}
					}
				}
			}
		}
	}

	return interactions
}

// CanPerformInteraction проверяет, может ли персонаж выполнить взаимодействие
func (g *Game) CanPerformInteraction(char *worldpkg.Character, interaction config.Interaction) bool {
	// Проверяем инструмент
	if interaction.Tool == "hand" {
		return char.HandsFree
	}

	// Проверяем, есть ли нужный инструмент в экипировке
	if equippedID, ok := char.Equipped[interaction.Tool]; ok {
		itemConfig := g.GetItemConfig(equippedID)
		return itemConfig != nil
	}

	return false
}

// AddToInventory добавляет предмет в инвентарь персонажа
func (g *Game) AddToInventory(char *worldpkg.Character, itemID int, count int) {
	if char.Inventory == nil {
		char.Inventory = make(map[int]worldpkg.InventoryItem)
	}

	// Ищем слот с таким же предметом
	for slotID, item := range char.Inventory {
		if item.ItemID == itemID {
			itemConfig := g.GetItemConfig(itemID)
			if itemConfig != nil && item.Count+count <= itemConfig.StackSize {
				item.Count += count
				char.Inventory[slotID] = item
				fmt.Printf("Добавлено %d x %s в инвентарь %s\n", count, itemConfig.Name, char.Name)
				return
			}
		}
	}

	// Ищем свободный слот
	for slotID := 0; slotID < 20; slotID++ {
		if _, exists := char.Inventory[slotID]; !exists {
			itemConfig := g.GetItemConfig(itemID)
			if itemConfig != nil {
				char.Inventory[slotID] = worldpkg.InventoryItem{
					ItemID: itemID,
					Count:  count,
				}
				fmt.Printf("Добавлено %d x %s в слот %d инвентаря %s\n", count, itemConfig.Name, slotID, char.Name)
				return
			}
		}
	}

	fmt.Printf("Инвентарь %s полон!\n", char.Name)
}

// PerformInteraction выполняет взаимодействие с объектом
func (g *Game) PerformInteraction(char *worldpkg.Character, objectID int, interaction config.Interaction) bool {
	// Находим объект
	obj := g.GetObjectAtPosition(char.Location, int(char.X+0.5))
	if obj == nil || obj.ID != objectID {
		// Проверяем соседние клетки
		pos := int(char.X + 0.5)
		neighborPositions := []int{pos - 1, pos, pos + 1}
		for _, neighborPos := range neighborPositions {
			if neighborPos >= 0 && neighborPos < len(g.State.LocationStates[char.Location].Road) {
				if tempObj := g.GetObjectAtPosition(char.Location, neighborPos); tempObj != nil && tempObj.ID == objectID {
					obj = tempObj
					break
				}
			}
		}
	}

	if obj == nil {
		fmt.Printf("Объект с ID %d не найден\n", objectID)
		return false
	}

	// Проверяем возможность выполнения
	if !g.CanPerformInteraction(char, interaction) {
		fmt.Printf("%s не может выполнить это действие. Нужен инструмент: %s\n", char.Name, interaction.Tool)
		return false
	}

	// Проверяем прочность объекта
	objConfig := g.GetObjectConfig(obj.TypeID)
	if objConfig == nil {
		return false
	}

	// Выполняем взаимодействие
	fmt.Printf("%s выполняет действие '%s' с %s...\n", char.Name, interaction.Type, objConfig.Name)

	// Добавляем предметы в инвентарь
	for _, result := range interaction.Results {
		g.AddToInventory(char, result.ItemID, result.Count)
	}

	// Определяем сколько прочности отнимать
	reduceDurability := interaction.ReduceDurability
	if reduceDurability == 0 {
		reduceDurability = 1 // Значение по умолчанию
	}

	// Уменьшаем прочность объекта
	obj.Durability -= reduceDurability

	// Проверяем, нужно ли превращать объект в другой тип
	if interaction.TransformTo > 0 && obj.Durability <= 0 {
		// Превращаем объект
		oldTypeID := obj.TypeID
		obj.TypeID = interaction.TransformTo
		obj.Durability = objConfig.MaxDurability // Сбрасываем прочность для нового объекта

		// Обновляем конфиг
		newObjConfig := g.GetObjectConfig(interaction.TransformTo)
		if newObjConfig != nil {
			fmt.Printf("%s превратился в %s!\n", objConfig.Name, newObjConfig.Name)
		}

		// Обновляем слой отображения
		g.UpdateObjectLayer(obj.LocationID, obj.X, oldTypeID, interaction.TransformTo)
	} else if interaction.DestroyOnComplete && obj.Durability <= 0 {
		// Удаляем объект
		g.RemoveObject(obj.ID)
		fmt.Printf("%s уничтожен!\n", objConfig.Name)
	} else if obj.Durability <= 0 {
		// Если объект должен быть уничтожен, но не указано явно
		g.RemoveObject(obj.ID)
		fmt.Printf("%s уничтожен!\n", objConfig.Name)
	} else {
		// Объект еще жив, но прочность уменьшилась
		fmt.Printf("%s: прочность %d/%d\n", objConfig.Name, obj.Durability, objConfig.MaxDurability)

		// Если это куст малины и прочность <= 0, превращаем в пустой куст
		if obj.TypeID == 5 && obj.Durability <= 0 { // 5 - raspberry_bush
			obj.TypeID = 9 // 9 - raspberry_bush_empty
			obj.Durability = g.GetObjectConfig(9).MaxDurability
			fmt.Printf("Куст малины опустел. Ягоды нужно ждать %d секунд.\n", g.GetObjectConfig(9).GrowthTime)
		}
	}

	return true
}

// UpdateObjectLayer обновляет слой отображения объекта
func (g *Game) UpdateObjectLayer(locationID int, pos int, oldTypeID int, newTypeID int) {
	locState := g.State.LocationStates[locationID]
	if locState == nil {
		return
	}

	// Обновляем слой в зависимости от типа объекта
	objConfig := g.GetObjectConfig(newTypeID)
	if objConfig == nil {
		return
	}

	// Определяем, в каком слое должен быть объект
	if objConfig.Foreground && pos >= 0 && pos < len(locState.Foreground) {
		locState.Foreground[pos] = newTypeID
	} else if objConfig.Background && pos >= 0 && pos < len(locState.Background) {
		locState.Background[pos] = newTypeID
	}
}

// RemoveObject удаляет объект из мира
func (g *Game) RemoveObject(objectID int) {
	// Находим объект чтобы узнать его позицию и тип
	var obj *worldpkg.WorldObject
	for _, tempObj := range g.GameWorld.Objects {
		if tempObj.ID == objectID {
			obj = tempObj
			break
		}
	}

	if obj == nil {
		return
	}

	// Получаем конфиг объекта
	objConfig := g.GetObjectConfig(obj.TypeID)

	// Обновляем слои отображения
	if objConfig != nil {
		locState := g.State.LocationStates[obj.LocationID]
		if locState != nil {
			pos := obj.X
			if objConfig.Foreground && pos >= 0 && int(pos) < len(locState.Foreground) {
				locState.Foreground[int(pos)] = 0
			}
			if objConfig.Background && pos >= 0 && int(pos) < len(locState.Background) {
				locState.Background[int(pos)] = 0
			}
		}
	}

	// Удаляем из мира
	delete(g.GameWorld.Objects, objectID)

	// Удаляем из локаций
	for _, loc := range g.GameWorld.Locations {
		delete(loc.Objects, objectID)
	}

	// Удаляем из состояния игры
	for locID, objects := range g.State.ObjectsByLocation {
		for i, tempObj := range objects {
			if tempObj.ID == objectID {
				g.State.ObjectsByLocation[locID] = append(objects[:i], objects[i+1:]...)
				break
			}
		}
	}
}

func (g *Game) UpdateCharacter(char *worldpkg.Character, elapsed float64) bool {
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

			// Освобождаем старую позицию
			if oldPos >= 0 && oldPos < len(locState.Foreground) {
				locState.Foreground[oldPos] = 0
			}

			// Занимаем новую позицию
			locState.Foreground[newPos] = char.ID

			// Применяем модификатор скорости дороги
			char.Speed = 0.7 * speedMod

			return true
		}
	}
	return false
}

func (g *Game) TryTransition(char *worldpkg.Character, side string) {
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

func (g *Game) removeCharFromSlice(slice []*worldpkg.Character, char *worldpkg.Character) []*worldpkg.Character {
	for i, c := range slice {
		if c.ID == char.ID {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func (g *Game) GetLocation(id int) *worldpkg.Location {
	for _, loc := range g.GameWorld.Locations {
		if loc.ID == id {
			return loc
		}
	}
	return nil
}

func (g *Game) GetPlayerCharacter() *worldpkg.Character {
	for _, char := range g.GameWorld.Characters {
		if char.Controlled == g.GameWorld.PlayerID {
			return char
		}
	}
	return nil
}

// UpdateCreature обновляет состояние существа
func (g *Game) UpdateCreature(creature *worldpkg.Creature, elapsed float64) {
	// Увеличиваем голод и жажду со временем
	creature.Hunger = min(100, creature.Hunger+1)
	creature.Thirst = min(100, creature.Thirst+1)

	creature.LastUpdate = time.Now()

	// Если у существа нет поведения, устанавливаем поведение по умолчанию
	if creature.CurrentBehavior == nil {
		g.SetDefaultBehavior(creature)
		return
	}

	// Проверяем, не завершилось ли текущее поведение
	if time.Since(creature.CurrentBehavior.StartTime).Seconds() >= creature.CurrentBehavior.Duration {
		// Поведение завершено, выбираем следующее
		g.ChooseNextBehavior(creature)
		return
	}

	// Выполняем текущее поведение
	switch creature.CurrentBehavior.Type {
	case "wander":
		g.ExecuteWanderBehavior(creature, elapsed)
	case "eat":
		g.ExecuteEatBehavior(creature, elapsed)
	case "rest":
		g.ExecuteRestBehavior(creature, elapsed)
	case "walk":
		g.ExecuteWalkBehavior(creature, elapsed)
	}
}

// SetDefaultBehavior устанавливает поведение по умолчанию
func (g *Game) SetDefaultBehavior(creature *worldpkg.Creature) {
	creatureConfig := g.GetCreatureConfig(creature.TypeID)
	if creatureConfig == nil {
		return
	}

	creature.CurrentBehavior = &worldpkg.CreatureBehavior{
		Type:      creatureConfig.DefaultBehavior,
		TargetPos: -1,
		StartTime: time.Now(),
		Duration:  g.GetBehaviorDuration(creatureConfig.DefaultBehavior),
	}
}

// ChooseNextBehavior выбирает следующее поведение для существа
func (g *Game) ChooseNextBehavior(creature *worldpkg.Creature) {
	creatureConfig := g.GetCreatureConfig(creature.TypeID)
	if creatureConfig == nil {
		return
	}

	// Проверяем голод
	if creature.Hunger > 70 && contains(creatureConfig.Behaviors, "eat") {
		// Пытаемся найти еду
		if g.FindFoodNearby(creature) {
			creature.CurrentBehavior = &worldpkg.CreatureBehavior{
				Type:      "eat",
				TargetPos: -1,
				StartTime: time.Now(),
				Duration:  5.0, // 5 секунд на еду
			}
			return
		}
	}

	// Выбираем случайное поведение из доступных
	availableBehaviors := creatureConfig.Behaviors
	if len(availableBehaviors) == 0 {
		availableBehaviors = []string{creatureConfig.DefaultBehavior}
	}

	randomBehavior := availableBehaviors[g.RandomInt(0, len(availableBehaviors)-1)]

	creature.CurrentBehavior = &worldpkg.CreatureBehavior{
		Type:      randomBehavior,
		TargetPos: -1,
		StartTime: time.Now(),
		Duration:  g.GetBehaviorDuration(randomBehavior),
	}

	// Для блуждания и ходьбы устанавливаем цель
	if randomBehavior == "wander" || randomBehavior == "walk" {
		g.SetMovementTarget(creature)
	}
}

// GetBehaviorDuration возвращает длительность поведения
func (g *Game) GetBehaviorDuration(behaviorType string) float64 {
	durations := map[string]float64{
		"wander": float64(g.RandomInt(3, 15)), // 3-15 секунд
		"eat":    5.0,                         // 5 секунд
		"rest":   float64(g.RandomInt(5, 30)), // 5-30 секунд
		"walk":   float64(g.RandomInt(2, 10)), // 2-10 секунд
		"attack": 3.0,                         // 3 секунды
		"flee":   5.0,                         // 5 секунд
	}

	if duration, ok := durations[behaviorType]; ok {
		return duration
	}
	return 5.0 // По умолчанию 5 секунд
}

// ExecuteWanderBehavior выполняет блуждающее поведение
func (g *Game) ExecuteWanderBehavior(creature *worldpkg.Creature, elapsed float64) {
	if creature.CurrentBehavior.TargetPos == -1 {
		g.SetMovementTarget(creature)
	}

	// Двигаемся к цели
	g.MoveCreatureToTarget(creature, elapsed)
}

// ExecuteEatBehavior выполняет поведение еды
func (g *Game) ExecuteEatBehavior(creature *worldpkg.Creature, elapsed float64) {
	// Проверяем, есть ли еда на текущей позиции
	pos := int(creature.X + 0.5)
	obj := g.GetObjectAtPosition(creature.Location, pos)
	if obj != nil {
		objConfig := g.GetObjectConfig(obj.TypeID)
		if objConfig != nil {
			// Проверяем, является ли объект едой для этого существа
			creatureConfig := g.GetCreatureConfig(creature.TypeID)
			if creatureConfig != nil && g.IsEdibleForCreature(objConfig.ID, creatureConfig.FavoriteFoods) {
				// Уменьшаем голод
				creature.Hunger = max(0, creature.Hunger-20)
				fmt.Printf("%s ест %s. Голод: %d\n", creature.Name, objConfig.Name, creature.Hunger)

				// Уменьшаем прочность объекта (съедаем его)
				obj.Durability -= 10
				if obj.Durability <= 0 {
					g.RemoveObject(obj.ID)
				}
			}
		}
	}
}

// ExecuteRestBehavior выполняет поведение отдыха
func (g *Game) ExecuteRestBehavior(creature *worldpkg.Creature, elapsed float64) {
	// Восстанавливаем здоровье во время отдыха
	if creature.Health < creature.MaxHealth {
		creature.Health = min(creature.MaxHealth, creature.Health+1)
	}
}

// ExecuteWalkBehavior выполняет поведение ходьбы
func (g *Game) ExecuteWalkBehavior(creature *worldpkg.Creature, elapsed float64) {
	if creature.CurrentBehavior.TargetPos == -1 {
		g.SetMovementTarget(creature)
	}

	// Двигаемся к цели
	g.MoveCreatureToTarget(creature, elapsed)
}

// SetMovementTarget устанавливает цель движения для существа
func (g *Game) SetMovementTarget(creature *worldpkg.Creature) {
	locState := g.State.LocationStates[creature.Location]
	if locState == nil || len(locState.Road) == 0 {
		creature.CurrentBehavior.TargetPos = -1
		return
	}

	// Выбираем случайную позицию в пределах 1-10 клеток от текущей
	currentPos := int(creature.X + 0.5)
	maxDistance := min(10, len(locState.Road)-1)
	distance := g.RandomInt(1, maxDistance)
	direction := 1
	if g.RandomInt(0, 1) == 0 {
		direction = -1
	}

	targetPos := currentPos + (direction * distance)

	// Проверяем границы
	if targetPos < 0 {
		targetPos = 0
	} else if targetPos >= len(locState.Road) {
		targetPos = len(locState.Road) - 1
	}

	// Проверяем доступность клетки
	if !g.IsPositionWalkable(creature.Location, targetPos) {
		// Пробуем другую позицию
		g.SetMovementTarget(creature)
		return
	}

	creature.CurrentBehavior.TargetPos = targetPos
}

// MoveCreatureToTarget двигает существо к цели
func (g *Game) MoveCreatureToTarget(creature *worldpkg.Creature, elapsed float64) {
	if creature.CurrentBehavior.TargetPos == -1 {
		return
	}

	currentPos := int(creature.X + 0.5)
	targetPos := creature.CurrentBehavior.TargetPos

	if currentPos == targetPos {
		// Достигли цели
		return
	}

	// Определяем направление
	direction := 1
	if targetPos < currentPos {
		direction = -1
	}

	// Проверяем следующую клетку
	nextPos := currentPos + direction
	if !g.IsPositionWalkable(creature.Location, nextPos) {
		// Клетка недоступна, выбираем новую цель
		g.SetMovementTarget(creature)
		return
	}

	// Двигаем существо
	creatureConfig := g.GetCreatureConfig(creature.TypeID)
	if creatureConfig == nil {
		return
	}

	creature.X += float64(direction) * creatureConfig.Speed * elapsed

	// Обновляем позицию в слое
	locState := g.State.LocationStates[creature.Location]
	if locState != nil && currentPos >= 0 && currentPos < len(locState.Foreground) {
		// Очищаем старую позицию
		locState.Foreground[currentPos] = 0
		// Занимаем новую позицию
		newPos := int(creature.X + 0.5)
		if newPos >= 0 && newPos < len(locState.Foreground) {
			locState.Foreground[newPos] = -creature.ID // Отрицательные ID для существ
		}
	}
}

// IsPositionWalkable проверяет, доступна ли позиция для движения
func (g *Game) IsPositionWalkable(locationID int, pos int) bool {
	if pos < 0 {
		return false
	}

	locState := g.State.LocationStates[locationID]
	if locState == nil || pos >= len(locState.Road) {
		return false
	}

	// Проверяем дорогу
	roadID := locState.Road[pos]
	if roadID == -1 {
		return false
	}

	// Проверяем землю
	groundID := locState.Ground[pos]
	groundConfig := g.GetGroundConfig(groundID)
	if groundConfig != nil && !groundConfig.Walkable {
		return false
	}

	// Проверяем, занята ли позиция другим существом или персонажем
	if locState.Foreground[pos] != 0 {
		return false
	}

	return true
}

// FindFoodNearby ищет еду рядом с существом
func (g *Game) FindFoodNearby(creature *worldpkg.Creature) bool {
	currentPos := int(creature.X + 0.5)
	creatureConfig := g.GetCreatureConfig(creature.TypeID)
	if creatureConfig == nil {
		return false
	}

	// Проверяем текущую позицию
	if obj := g.GetObjectAtPosition(creature.Location, currentPos); obj != nil {
		if g.IsEdibleForCreature(obj.TypeID, creatureConfig.FavoriteFoods) {
			return true
		}
	}

	// Проверяем соседние клетки (в радиусе 3 клеток)
	for distance := 1; distance <= 3; distance++ {
		leftPos := currentPos - distance
		rightPos := currentPos + distance

		if leftPos >= 0 {
			if obj := g.GetObjectAtPosition(creature.Location, leftPos); obj != nil {
				if g.IsEdibleForCreature(obj.TypeID, creatureConfig.FavoriteFoods) {
					// Устанавливаем цель движения к еде
					creature.CurrentBehavior.TargetPos = leftPos
					return true
				}
			}
		}

		if rightPos < len(g.State.LocationStates[creature.Location].Road) {
			if obj := g.GetObjectAtPosition(creature.Location, rightPos); obj != nil {
				if g.IsEdibleForCreature(obj.TypeID, creatureConfig.FavoriteFoods) {
					creature.CurrentBehavior.TargetPos = rightPos
					return true
				}
			}
		}
	}

	return false
}

// IsEdibleForCreature проверяет, съедобен ли объект для существа
func (g *Game) IsEdibleForCreature(objectTypeID int, favoriteFoods []int) bool {
	for _, foodID := range favoriteFoods {
		if objectTypeID == foodID {
			return true
		}
	}
	return false
}

// RandomInt возвращает случайное целое число в диапазоне [min, max]
func (g *Game) RandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

// PrintInventory выводит инвентарь персонажа
func (g *Game) PrintInventory(char *worldpkg.Character) {
	fmt.Printf("\n=== ИНВЕНТАРЬ %s ===\n", char.Name)

	if len(char.Inventory) == 0 {
		fmt.Println("Инвентарь пуст")
		return
	}

	totalWeight := 0.0
	for slotID, item := range char.Inventory {
		itemConfig := g.GetItemConfig(item.ItemID)
		if itemConfig != nil {
			slotWeight := float64(item.Count) * itemConfig.Weight
			totalWeight += slotWeight
			fmt.Printf("Слот %d: %d x %s (%.2f кг)\n",
				slotID, item.Count, itemConfig.Name, slotWeight)
		}
	}

	fmt.Printf("Общий вес: %.2f кг\n", totalWeight)

	// Экипировка
	if len(char.Equipped) > 0 {
		fmt.Println("\nЭкипировка:")
		for toolType, itemID := range char.Equipped {
			itemConfig := g.GetItemConfig(itemID)
			if itemConfig != nil {
				fmt.Printf("  %s: %s\n", toolType, itemConfig.Name)
			}
		}
	}

	fmt.Printf("Руки свободны: %v\n", char.HandsFree)
}

// PrintAvailableInteractions выводит доступные взаимодействия
func (g *Game) PrintAvailableInteractions(char *worldpkg.Character) {
	interactions := g.GetAvailableInteractions(char)

	if len(interactions) == 0 {
		fmt.Println("Нет доступных взаимодействий")
		return
	}

	fmt.Printf("\n=== ДОСТУПНЫЕ ВЗАИМОДЕЙСТВИЯ ДЛЯ %s ===\n", char.Name)

	interactionIndex := 0
	pos := int(char.X + 0.5)

	// Проверяем текущую позицию
	if obj := g.GetObjectAtPosition(char.Location, pos); obj != nil {
		objConfig := g.GetObjectConfig(obj.TypeID)
		if objConfig != nil {
			fmt.Printf("\nОбъект на позиции %d: %s (ID: %d) прочность: %d/%d\n",
				pos, objConfig.Name, obj.ID, obj.Durability, objConfig.MaxDurability)
			for _, interaction := range objConfig.Interactions {
				if g.CanPerformInteraction(char, interaction) {
					fmt.Printf("  [%d] %s (инструмент: %s, время: %dс)\n",
						interactionIndex, interaction.Type, interaction.Tool, interaction.Time)

					// Показываем эффекты
					if interaction.ReduceDurability > 0 {
						fmt.Printf("      отнимает прочность: %d\n", interaction.ReduceDurability)
					}
					if interaction.TransformTo > 0 {
						newObjConfig := g.GetObjectConfig(interaction.TransformTo)
						if newObjConfig != nil {
							fmt.Printf("      превращается в: %s\n", newObjConfig.Name)
						}
					}
					if interaction.DestroyOnComplete {
						fmt.Printf("      уничтожает объект\n")
					}

					// Показываем награды
					for _, result := range interaction.Results {
						itemConfig := g.GetItemConfig(result.ItemID)
						if itemConfig != nil {
							fmt.Printf("      -> %d x %s\n", result.Count, itemConfig.Name)
						}
					}
					interactionIndex++
				}
			}
		}
	}

	// Проверяем соседние клетки
	neighborPositions := []int{pos - 1, pos + 1}
	for _, neighborPos := range neighborPositions {
		if neighborPos >= 0 && neighborPos < len(g.State.LocationStates[char.Location].Road) {
			if obj := g.GetObjectAtPosition(char.Location, neighborPos); obj != nil {
				objConfig := g.GetObjectConfig(obj.TypeID)
				if objConfig != nil {
					fmt.Printf("\nОбъект на позиции %d: %s (ID: %d) прочность: %d/%d\n",
						neighborPos, objConfig.Name, obj.ID, obj.Durability, objConfig.MaxDurability)
					for _, interaction := range objConfig.Interactions {
						if g.CanPerformInteraction(char, interaction) {
							fmt.Printf("  [%d] %s (инструмент: %s, время: %dс)\n",
								interactionIndex, interaction.Type, interaction.Tool, interaction.Time)

							// Показываем эффекты
							if interaction.ReduceDurability > 0 {
								fmt.Printf("      отнимает прочность: %d\n", interaction.ReduceDurability)
							}
							if interaction.TransformTo > 0 {
								newObjConfig := g.GetObjectConfig(interaction.TransformTo)
								if newObjConfig != nil {
									fmt.Printf("      превращается в: %s\n", newObjConfig.Name)
								}
							}
							if interaction.DestroyOnComplete {
								fmt.Printf("      уничтожает объект\n")
							}

							// Показываем награды
							for _, result := range interaction.Results {
								itemConfig := g.GetItemConfig(result.ItemID)
								if itemConfig != nil {
									fmt.Printf("      -> %d x %s\n", result.Count, itemConfig.Name)
								}
							}
							interactionIndex++
						}
					}
				}
			}
		}
	}

	fmt.Println("\nДля выполнения действия введите: act <ID объекта> <номер действия>")
}

// PerformInteractionByIndex выполняет взаимодействие по индексу
func (g *Game) PerformInteractionByIndex(char *worldpkg.Character, objectID int, interactionIndex int) {
	// Находим объект
	obj := g.GetObjectAtPosition(char.Location, int(char.X+0.5))
	if obj == nil || obj.ID != objectID {
		// Проверяем соседние клетки
		pos := int(char.X + 0.5)
		neighborPositions := []int{pos - 1, pos, pos + 1}
		for _, neighborPos := range neighborPositions {
			if neighborPos >= 0 && neighborPos < len(g.State.LocationStates[char.Location].Road) {
				if tempObj := g.GetObjectAtPosition(char.Location, neighborPos); tempObj != nil && tempObj.ID == objectID {
					obj = tempObj
					break
				}
			}
		}
	}

	if obj == nil {
		fmt.Printf("Объект с ID %d не найден\n", objectID)
		return
	}

	objConfig := g.GetObjectConfig(obj.TypeID)
	if objConfig == nil {
		fmt.Printf("Конфигурация объекта не найдена\n")
		return
	}

	// Находим взаимодействие по индексу
	index := 0
	for _, interaction := range objConfig.Interactions {
		if g.CanPerformInteraction(char, interaction) {
			if index == interactionIndex {
				g.PerformInteraction(char, objectID, interaction)
				return
			}
			index++
		}
	}

	fmt.Printf("Действие с индексом %d не найдено или недоступно\n", interactionIndex)
}

// UpdateWorldObjects обновляет состояние объектов мира
func (g *Game) UpdateWorldObjects(elapsed float64) {
	// В будущем здесь будет логика роста объектов
	// Например, проверка growth_time и обновление growth_stage
}

func (g *Game) PrintState() {
	fmt.Println("\n=== СОСТОЯНИЕ МИРА ===")
	fmt.Printf("ID игрока: %d\n", g.GameWorld.PlayerID)

	for _, loc := range g.GameWorld.Locations {
		fmt.Printf("\nЛокация %d: %s\n", loc.ID, loc.Name)
		locState := g.State.LocationStates[loc.ID]

		// Выводим задний фон
		fmt.Print("Задний фон:   [")
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

		// Выводим дорожный слой с персонажами и существами
		fmt.Print("Дорога:       [")
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
					// Проверяем, существо ли это (отрицательный ID)
					if locState.Foreground[i] < 0 {
						creatureID := -locState.Foreground[i]
						if creature := g.GetCreatureByID(creatureID); creature != nil {
							creatureConfig := g.GetCreatureConfig(creature.TypeID)
							if creatureConfig != nil {
								fmt.Printf("%c", strings.ToLower(creatureConfig.Name)[0])
							} else {
								fmt.Print("?")
							}
						}
					} else {
						// Это объект на переднем плане
						objConfig := g.GetObjectConfig(locState.Foreground[i])
						if objConfig != nil {
							fmt.Print(objConfig.Name[0:1])
						} else {
							fmt.Print("?")
						}
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
		fmt.Print("Земля:        [")
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

		// Выводим передний фон
		fmt.Print("Передний фон: [")
		for i := 0; i < len(locState.Foreground); i++ {
			if locState.Foreground[i] == 0 {
				fmt.Print(" ")
			} else {
				objConfig := g.GetObjectConfig(locState.Foreground[i])
				if objConfig != nil {
					fmt.Print(objConfig.Name[0:1]) // Первая буква названия
				} else {
					fmt.Print("?")
				}
			}
			if i < len(locState.Foreground)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Println("]")

		// Персонажи в этой локации
		if chars, ok := g.State.CharsByLocation[loc.ID]; ok && len(chars) > 0 {
			fmt.Println("Персонажи:")
			for _, char := range chars {
				controlStatus := "NPC"
				if char.Controlled == g.GameWorld.PlayerID {
					controlStatus = "ИГРОК"
				}
				fmt.Printf("  %s (ID: %d) поз: %.1f, напр: %d, верт: %d, скорость: %.1f [%s]\n",
					char.Name, char.ID, char.X, char.Direction, char.Vertical, char.Speed, controlStatus)
			}
		}

		// Существа в этой локации
		if creatures, ok := g.State.CreaturesByLocation[loc.ID]; ok && len(creatures) > 0 {
			fmt.Println("Существа:")
			for _, creature := range creatures {
				creatureConfig := g.GetCreatureConfig(creature.TypeID)
				if creatureConfig != nil {
					behaviorInfo := "бездействует"
					if creature.CurrentBehavior != nil {
						behaviorInfo = creature.CurrentBehavior.Type
						if creature.CurrentBehavior.TargetPos != -1 {
							behaviorInfo += fmt.Sprintf(" -> клетка %d", creature.CurrentBehavior.TargetPos)
						}
					}

					fmt.Printf("  %s (ID: %d) поз: %.1f, здоровье: %d/%d, голод: %d, поведение: %s\n",
						creatureConfig.Name, creature.ID, creature.X,
						creature.Health, creature.MaxHealth, creature.Hunger, behaviorInfo)
				}
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
		fmt.Printf("Типов объектов: %d, Типов дорог: %d, Типов земли: %d, Типов предметов: %d, Типов существ: %d\n",
			len(g.Registries.ObjectTypeByID),
			len(g.Registries.RoadTypeByID),
			len(g.Registries.GroundTypeByID),
			len(g.Registries.ItemTypeByID),
			len(g.Registries.CreatureTypeByID))
	}

	fmt.Println("\nКоманды: a/d - влево/вправо, w/s - вверх/вниз, stop - остановка, i - инвентарь, act - взаимодействия, x - состояние, save - сохранить, exit - выход")
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
	case "i":
		g.PrintInventory(playerChar)
	case "act":
		g.PrintAvailableInteractions(playerChar)
	case "x":
		g.PrintState()
	case "save":
		// Сохраняем только игровое состояние (без конфигов)
		saveWorld := &worldpkg.World{
			PlayerID:   g.GameWorld.PlayerID,
			Characters: g.GameWorld.Characters,
			Locations:  g.GameWorld.Locations,
			Objects:    g.GameWorld.Objects,
			Creatures:  g.GameWorld.Creatures,
		}
		if err := worldpkg.SaveWorld(saveWorld, "data/save/world.json"); err != nil {
			fmt.Printf("Ошибка сохранения: %v\n", err)
		} else {
			fmt.Println("Мир сохранен в data/save/world.json")
		}
	case "exit":
		g.State.Running = false
		g.ExitChan <- true
	default:
		// Пробуем выполнить действие формата "act <id> <index>"
		if strings.HasPrefix(input, "act ") {
			parts := strings.Split(input, " ")
			if len(parts) == 3 {
				var objectID, interactionIndex int
				if _, err := fmt.Sscanf(parts[1], "%d", &objectID); err == nil {
					if _, err := fmt.Sscanf(parts[2], "%d", &interactionIndex); err == nil {
						g.PerformInteractionByIndex(playerChar, objectID, interactionIndex)
						return
					}
				}
			}
		}
		fmt.Println("Неизвестная команда. Доступные: a, d, w, s, stop, i, act, x, save, exit, act <id> <index>")
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

			// Обновляем персонажей
			for _, char := range g.GameWorld.Characters {
				if g.UpdateCharacter(char, elapsed) {
					updated = true
				}
			}

			// Обновляем существ
			for _, creature := range g.GameWorld.Creatures {
				g.UpdateCreature(creature, elapsed)
				updated = true // Всегда обновляем, так как существа могут двигаться
			}

			// Обновляем объекты мира (рост, восстановление)
			g.UpdateWorldObjects(elapsed)

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

// contains проверяет наличие строки в срезе
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
