package debug

import (
	"LOIL-server/internal/world"
	"fmt"
)

// CommandHandler - обработчик отладочных команд
type CommandHandler struct {
	printer *Printer
}

func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		printer: NewPrinter(),
	}
}

// HandleDebugCommand обрабатывает отладочные команды
func (h *CommandHandler) HandleDebugCommand(cmd string, w *world.World, locationStates map[int]*LocationState, charsByLocation map[int][]*world.Character) bool {
	switch cmd {
	case "x":
		h.printer.PrintWorldState(w, locationStates, charsByLocation)
		return true
	case "help":
		h.PrintHelp()
		return true
	default:
		return false
	}
}

// PrintHelp выводит справку по отладочным командам
func (h *CommandHandler) PrintHelp() {
	fmt.Println("\n=== ОТЛАДОЧНЫЕ КОМАНДЫ ===")
	fmt.Println("x - вывести состояние мира")
	fmt.Println("help - показать эту справку")
	fmt.Println("exit - выйти из отладочного режима")
}
