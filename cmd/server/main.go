package main

import (
	"LOIL-server/internal/game"
	"LOIL-server/internal/world"
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
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

	// Запускаем игровой цикл в горутине
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
			// Даем время на завершение
			time.Sleep(100 * time.Millisecond)
			break
		}

		g.InputChan <- input

		// Небольшая задержка для вывода
		time.Sleep(50 * time.Millisecond)
	}
}
