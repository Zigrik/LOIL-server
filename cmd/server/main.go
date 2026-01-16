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

	// Запускаем обработчик ввода (чистые игровые команды)
	runGameInputHandler(g)

	fmt.Println("Игра завершена.")
}

func runGameInputHandler(g *game.Game) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== ИГРА ЗАПУЩЕНА ===")
	fmt.Println("Игровые команды: a/d - влево/вправо, w/s - вверх/вниз, stop - остановка, exit - выход")

	for {
		fmt.Print("\n> ")
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
