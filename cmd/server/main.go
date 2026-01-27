package main

import (
	"LOIL-server/internal/config"
	"LOIL-server/internal/game"
	"LOIL-server/internal/network"
	worldpkg "LOIL-server/internal/world"
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	// Парсим флаги
	serverAddr := flag.String("addr", ":8080", "Адрес WebSocket сервера")
	headless := flag.Bool("headless", false, "Запуск без интерактивной консоли")
	flag.Parse()

	// Загружаем конфигурации
	configs, err := config.LoadConfigs()
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигураций: %v\n", err)
		return
	}

	// Загружаем мир
	world, err := worldpkg.LoadWorld("data/world.json", configs)
	if err != nil {
		fmt.Printf("Ошибка загрузки мира: %v\n", err)
		return
	}

	if *headless {
		// Серверный режим с сетью
		runServerMode(world, *serverAddr)
	} else {
		// Консольный режим для отладки
		runConsoleMode(world)
	}
}

func runServerMode(w *worldpkg.World, addr string) {
	fmt.Printf("Запуск сервера на %s...\n", addr)

	// Создаем игру
	g := game.NewGame(w)
	g.Initialize()

	// Создаем мост между игрой и сетью
	bridge := game.NewGameNetworkBridge(g)

	// Настраиваем сервер
	serverConfig := &network.ServerConfig{
		Addr:           addr,
		UpdateInterval: 100 * time.Millisecond, // 10 FPS
		PingInterval:   30 * time.Second,
		MaxMessageSize: 1024 * 10, // 10KB
	}

	// Создаем и запускаем сервер
	server := network.NewServer(bridge, serverConfig)

	// Запускаем игровой цикл в отдельной горутине
	go g.RunGameLoop()

	// Запускаем сервер (блокирующий вызов)
	if err := server.Start(); err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
	}
}

func runConsoleMode(w *worldpkg.World) {
	// Консольный режим без сети
	g := game.NewGame(w)
	g.Initialize()

	go g.RunGameLoop()
	runInputHandler(g)

	fmt.Println("Игра завершена.")
}

func runInputHandler(g *game.Game) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== КОНСОЛЬНЫЙ РЕЖИМ ===")
	fmt.Println("Команды: a/d - влево/вправо, w/s - вверх/вниз, stop - остановка")
	fmt.Println("         i - инвентарь, act - взаимодействия, x - состояние")
	fmt.Println("         save - сохранить, exit - выход")

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
