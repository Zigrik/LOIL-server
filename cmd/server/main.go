package main

import (
	"LOIL-server/internal/config"
	"LOIL-server/internal/game"
	worldpkg "LOIL-server/internal/world"
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	// Загружаем конфигурации
	configs, err := config.LoadConfigs()
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигураций: %v\n", err)
		return
	}

	// Загружаем мир с конфигами
	world, err := worldpkg.LoadWorld("data/world.json", configs)
	if err != nil {
		fmt.Printf("Ошибка загрузки мира: %v\n", err)
		return
	}

	// Создаем и инициализируем игру
	g := game.NewGame(world)
	g.Initialize()

	// Запускаем игровой цикл
	go g.RunGameLoop()

	// Запускаем обработчик ввода
	runInputHandler(g)

	fmt.Println("Игра завершена.")
}

func runInputHandler(g *game.Game) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== ИГРА ЗАПУЩЕНА ===")
	fmt.Println("Команды: a/d - влево/вправо, w/s - вверх/вниз, stop - остановка, x - состояние, save - сохранить, exit - выход")
	g.PrintState()

	for {
		fmt.Print("\nВведите команду: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		if input == "exit" {
			g.InputChan <- input
			time.Sleep(100 * time.Millisecond)
			break
		}

		g.InputChan <- input
		time.Sleep(50 * time.Millisecond)
	}
}
