package debug

import (
	"LOIL-server/internal/world"
	"fmt"
)

// Printer - структура для отладочного вывода
type Printer struct{}

func NewPrinter() *Printer {
	return &Printer{}
}

// PrintWorldState выводит состояние мира
func (p *Printer) PrintWorldState(w *world.World, locationStates map[int]*LocationState, charsByLocation map[int][]*world.Character) {
	fmt.Println("\n=== СОСТОЯНИЕ МИРА ===")
	fmt.Printf("ID игрока: %d\n", w.PlayerID)

	for _, loc := range w.Locations {
		fmt.Printf("\nЛокация %d: %s\n", loc.ID, loc.Name)
		locState := locationStates[loc.ID]

		// Выводим дорожный слой с персонажами
		fmt.Println("Дорожный слой с персонажами:")
		fmt.Print("Дорога: [")
		for i := 0; i < len(locState.Road); i++ {
			if locState.Foreground[i] != 0 {
				for _, char := range charsByLocation[loc.ID] {
					if char.ID == locState.Foreground[i] && int(char.X+0.5) == i {
						fmt.Printf("%c", char.Name[0])
						break
					}
				}
			} else if locState.Road[i] == -1 {
				fmt.Print("#") // Нет дороги
			} else if locState.Road[i] >= 0 {
				fmt.Print(".") // Есть дорога
			}
			if i < len(locState.Road)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Println("]")

		// Выводим слой земли
		fmt.Print("Земля:  [")
		for i := 0; i < len(locState.Ground); i++ {
			switch locState.Ground[i] {
			case -2:
				fmt.Print("≈") // Река
			case -1:
				fmt.Print("~") // Ручей
			case 0:
				fmt.Print(" ") // Земля
			case 1:
				fmt.Print(".") // Песок
			case 2:
				fmt.Print("□") // Глина
			case 3:
				fmt.Print("■") // Камень
			default:
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
			switch locState.Background[i] {
			case 0:
				fmt.Print(" ")
			case 1:
				fmt.Print("T") // Дерево
			case 2:
				fmt.Print("H") // Дом
			case 3:
				fmt.Print("F") // Забор
			default:
				fmt.Print("?")
			}
			if i < len(locState.Background)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Println("]")

		// Персонажи в этой локации
		if chars, ok := charsByLocation[loc.ID]; ok && len(chars) > 0 {
			fmt.Println("Персонажи:")
			for _, char := range chars {
				// Определяем статус управления
				var controlStatus string
				if char.Controlled == w.PlayerID {
					controlStatus = "ИГРОК"
				} else {
					controlStatus = "NPC"
				}

				fmt.Printf("  %s (ID: %d) поз: %.1f, напр: %d, верт: %d, скорость: %.1f [%s]\n",
					char.Name, char.ID, char.X, char.Direction, char.Vertical, char.Speed, controlStatus)
			}
		}
	}
	fmt.Println("\nКоманды: a/d - влево/вправо, w/s - вверх/вниз, stop - остановка, x - состояние, save - сохранить, exit - выход")
}

// PrintLocation выводит конкретную локацию
func (p *Printer) PrintLocation(loc *world.Location, locState *LocationState, chars []*world.Character) {
	fmt.Printf("\nЛокация %d: %s\n", loc.ID, loc.Name)

	// Выводим все слои
	p.PrintLayers(locState)

	// Выводим персонажей
	if len(chars) > 0 {
		fmt.Println("Персонажи:")
		for _, char := range chars {
			fmt.Printf("  %s (ID: %d) позиция: %.1f\n", char.Name, char.ID, char.X)
		}
	}
}

// PrintLayers выводит все слои локации
func (p *Printer) PrintLayers(locState *LocationState) {
	fmt.Println("Передний фон (объекты):")
	p.printLayer(locState.Foreground)

	fmt.Println("Дорожный слой:")
	p.printLayer(locState.Road)

	fmt.Println("Слой земли:")
	p.printGroundLayer(locState.Ground)

	fmt.Println("Задний фон:")
	p.printBackgroundLayer(locState.Background)
}

func (p *Printer) printLayer(layer []int) {
	fmt.Print("[")
	for i, val := range layer {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%d", val)
	}
	fmt.Println("]")
}

func (p *Printer) printGroundLayer(layer []int) {
	fmt.Print("[")
	for i, val := range layer {
		if i > 0 {
			fmt.Print(" ")
		}
		switch val {
		case -2:
			fmt.Print("≈") // Река
		case -1:
			fmt.Print("~") // Ручей
		case 0:
			fmt.Print(" ") // Земля
		case 1:
			fmt.Print(".") // Песок
		case 2:
			fmt.Print("□") // Глина
		case 3:
			fmt.Print("■") // Камень
		default:
			fmt.Printf("%d", val)
		}
	}
	fmt.Println("]")
}

func (p *Printer) printBackgroundLayer(layer []int) {
	fmt.Print("[")
	for i, val := range layer {
		if i > 0 {
			fmt.Print(" ")
		}
		switch val {
		case 0:
			fmt.Print(" ")
		case 1:
			fmt.Print("T") // Дерево
		case 2:
			fmt.Print("H") // Дом
		case 3:
			fmt.Print("F") // Забор
		default:
			fmt.Printf("%d", val)
		}
	}
	fmt.Println("]")
}

// LocationState - копия из game.go для отладочного вывода
type LocationState struct {
	Foreground []int
	Road       []int
	Ground     []int
	Background []int
}
