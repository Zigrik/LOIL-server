package main

import (
	"fmt"
	"time"
)

type Object struct {
	name string
	x    float64
}

func main() {
	// Объявляем слайс
	area := make([]int, 20)

	// Создаем объект
	bob := Object{
		name: "Bob",
		x:    0.0,
	}

	// Устанавливаем начальное положение
	currentPos := 0
	area[currentPos] = 1

	// Направление движения: 1 - вперед, -1 - назад
	direction := 1

	// Время ожидания на краях (в секундах)
	waitTime := 3 * time.Second

	// Скорость движения (единиц в секунду)
	speed := 0.7

	// Время последнего обновления
	lastUpdate := time.Now()

	// Флаг для определения, стоит ли объект на краю
	isWaiting := false
	waitStart := time.Now()

	// Выводим начальное состояние
	printArea(area, bob)

	// Бесконечный цикл
	for {
		currentTime := time.Now()
		elapsed := currentTime.Sub(lastUpdate).Seconds()

		if isWaiting {
			// Проверяем, закончилось ли время ожидания
			if currentTime.Sub(waitStart) >= waitTime {
				isWaiting = false
				direction *= -1 // Меняем направление
				lastUpdate = currentTime
			}
		} else {
			// Вычисляем новую позицию
			bob.x += float64(direction) * speed * elapsed

			// Проверяем границы
			if direction == 1 && bob.x >= 19 {
				bob.x = 19
				isWaiting = true
				waitStart = currentTime
			} else if direction == -1 && bob.x <= 0 {
				bob.x = 0
				isWaiting = true
				waitStart = currentTime
			}

			// Определяем новую целочисленную позицию (округление)
			newPos := int(bob.x + 0.5)

			// Проверяем, изменилась ли позиция в слайсе
			if newPos != currentPos && newPos >= 0 && newPos < len(area) {
				// Очищаем старую позицию
				area[currentPos] = 0

				// Устанавливаем новую позицию
				area[newPos] = 1
				currentPos = newPos

				// Выводим обновленное состояние
				printArea(area, bob)
			}

			lastUpdate = currentTime
		}

		// Небольшая задержка для уменьшения нагрузки на CPU
		time.Sleep(50 * time.Millisecond)
	}
}

// Функция для вывода слайса в консоль
func printArea(area []int, obj Object) {
	fmt.Printf("\nПозиция %s: %.1f\n", obj.name, obj.x)
	fmt.Print("[")
	for i, val := range area {
		if val == 1 {
			fmt.Print("B")
		} else {
			fmt.Print(".")
		}
		if i < len(area)-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println("]")
}
