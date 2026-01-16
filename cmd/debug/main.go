package main

import (
	"LOIL-server/internal/debug"
	"LOIL-server/internal/game"
	"LOIL-server/internal/world"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Загружаем мир
	w, err := world.LoadWorld("world.json")
	if err != nil {
		fmt.Printf("Ошибка загрузки мира: %v\n", err)
		return
	}

	// Создаем и инициализируем игру
	g := game.NewGame(w)
	g.Initialize()

	// Создаем отладочный обработчик
	debugHandler := debug.NewCommandHandler()

	// Запускаем отладочный интерфейс
	runDebugInterface(g, debugHandler)
}

func runDebugInterface(g *game.Game, debugHandler *debug.CommandHandler) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== ОТЛАДОЧНЫЙ РЕЖИМ ===")
	fmt.Println("Для получения списка команд введите 'help'")

	for {
		fmt.Print("\ndebug> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		if input == "exit" {
			fmt.Println("Выход из отладочного режима")
			break
		}

		// Конвертируем game.LocationState в debug.LocationState
		locationStates := convertLocationStates(g.GetLocationStates())

		// Обрабатываем отладочные команды
		if debugHandler.HandleDebugCommand(
			input,
			g.GetWorld(),
			locationStates,
			g.GetCharsByLocation(),
		) {
			continue
		}

		// Если не отладочная команда, передаем в игровой цикл
		g.InputChan <- input
	}
}

// convertLocationStates конвертирует game.LocationState в debug.LocationState
func convertLocationStates(gameStates map[int]*game.LocationState) map[int]*debug.LocationState {
	result := make(map[int]*debug.LocationState)

	for id, state := range gameStates {
		result[id] = &debug.LocationState{
			Foreground: state.Foreground,
			Road:       state.Road,
			Ground:     state.Ground,
			Background: state.Background,
		}
	}

	return result
}
