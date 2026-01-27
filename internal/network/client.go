package network

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// writePump отправляет сообщения клиенту
func (c *Client) writePump() {
	ticker := time.NewTicker(c.Server.Config.PingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		c.Server.Unregister <- c
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Канал закрыт
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()

			if err != nil {
				log.Printf("Ошибка отправки клиенту %s: %v", c.Info.ID, err)
				return
			}

		case <-ticker.C:
			// Уже обрабатывается в sendPings сервера
		}
	}
}

// readPump читает сообщения от клиента
func (c *Client) readPump() {
	defer func() {
		c.Server.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.Server.Config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.Server.Config.ReadTimeout))

	// Устанавливаем pong handler
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.Server.Config.ReadTimeout))
		c.Info.LastActivity = time.Now()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка чтения от клиента %s: %v", c.Info.ID, err)
			}
			break
		}

		c.Info.LastActivity = time.Now()
		c.handleMessage(message)
	}
}

// handleMessage обрабатывает входящее сообщение
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError("invalid_format", "Неверный формат JSON")
		return
	}

	// Обрабатываем в зависимости от типа
	switch msg.Type {
	case MsgJoin:
		c.handleJoin(msg.Payload)
	case MsgMove:
		c.handleMove(msg.Payload)
	case MsgStop:
		c.handleStop()
	case MsgInteract:
		c.handleInteract(msg.Payload)
	case MsgPong:
		// Обновляем время последней активности
		c.Info.LastActivity = time.Now()
	default:
		c.sendError("unknown_type", "Неизвестный тип сообщения: "+string(msg.Type))
	}
}

// handleJoin обрабатывает присоединение к игре
func (c *Client) handleJoin(payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		c.sendError("parse_error", "Ошибка парсинга запроса")
		return
	}

	var req JoinRequest
	if err := json.Unmarshal(data, &req); err != nil {
		c.sendError("invalid_request", "Неверный формат запроса")
		return
	}

	// Проверяем обязательные поля
	if req.PlayerID == 0 || req.LocationID == 0 {
		c.sendError("missing_fields", "Не указаны player_id или location_id")
		return
	}

	// Сохраняем информацию о клиенте
	c.Info.PlayerID = req.PlayerID
	c.Info.CharacterID = req.CharacterID
	c.Info.LocationID = req.LocationID

	// Уведомляем игру о присоединении
	charState, err := c.Server.Game.HandleJoin(req.PlayerID, req.CharacterID, req.LocationID)
	if err != nil {
		c.sendError("join_failed", err.Error())
		return
	}

	// Устанавливаем CharacterID из ответа игры (если не был указан)
	if c.Info.CharacterID == 0 && charState != nil {
		c.Info.CharacterID = charState.ID
	}

	// Получаем полное состояние локации для клиента
	worldState := c.getFullWorldState(req.LocationID)
	if worldState == nil {
		c.sendError("location_not_found", "Локация не найдена")
		return
	}

	// Добавляем информацию о персонаже игрока
	if charState != nil {
		// Убедимся, что персонаж игрока есть в списке
		found := false
		for _, char := range worldState.Characters {
			if char.ID == charState.ID {
				found = true
				break
			}
		}
		if !found {
			worldState.Characters = append(worldState.Characters, charState)
		}
	}

	// Отправляем состояние клиенту
	msg := Message{
		Type:    MsgWorldState,
		Payload: worldState,
		Time:    Now(),
		Seq:     c.getNextSeq(),
	}

	c.sendMessage(msg)

	log.Printf("Клиент %s присоединился как игрок %d (персонаж %d) в локацию %d",
		c.Info.ID, req.PlayerID, c.Info.CharacterID, req.LocationID)
}

// getFullWorldState получает полное состояние мира для клиента
func (c *Client) getFullWorldState(locationID int) *WorldState {
	return &WorldState{
		PlayerID:   c.Info.PlayerID,
		Location:   c.Server.Game.GetLocationState(locationID),
		Characters: c.Server.Game.GetCharactersInLocation(locationID),
		Creatures:  c.Server.Game.GetCreaturesInLocation(locationID),
		Objects:    c.Server.Game.GetObjectsInLocation(locationID),
		ServerTime: c.Server.Game.GetServerTime(),
	}
}

// handleMove обрабатывает движение
func (c *Client) handleMove(payload interface{}) {
	if c.Info.PlayerID == 0 {
		c.sendError("not_joined", "Сначала нужно присоединиться к игре")
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		c.sendError("parse_error", "Ошибка парсинга запроса")
		return
	}

	var req MoveRequest
	if err := json.Unmarshal(data, &req); err != nil {
		c.sendError("invalid_request", "Неверный формат запроса движения")
		return
	}

	// Проверяем валидность направления
	if req.Direction < -1 || req.Direction > 1 {
		c.sendError("invalid_direction", "Направление должно быть -1, 0 или 1")
		return
	}
	if req.Vertical < -1 || req.Vertical > 1 {
		c.sendError("invalid_vertical", "Вертикаль должна быть -1, 0 или 1")
		return
	}

	// Передаем движение в игру
	if err := c.Server.Game.HandleMove(c.Info.PlayerID, req.Direction, req.Vertical); err != nil {
		c.sendError("move_failed", err.Error())
		return
	}

	// Отправляем подтверждение
	msg := Message{
		Type: MsgCharacterUpdate,
		Payload: CharacterUpdate{
			CharacterID: c.Info.CharacterID,
			State:       c.Server.Game.GetCharacterByID(c.Info.CharacterID),
			ServerTime:  Now(),
		},
		Time: Now(),
		Seq:  c.getNextSeq(),
	}

	c.sendMessage(msg)
}

// handleStop обрабатывает остановку
func (c *Client) handleStop() {
	if c.Info.PlayerID == 0 {
		return
	}

	if err := c.Server.Game.HandleStop(c.Info.PlayerID); err != nil {
		c.sendError("stop_failed", err.Error())
		return
	}
}

// handleInteract обрабатывает взаимодействие
func (c *Client) handleInteract(payload interface{}) {
	if c.Info.PlayerID == 0 {
		c.sendError("not_joined", "Сначала нужно присоединиться к игре")
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		c.sendError("parse_error", "Ошибка парсинга запроса")
		return
	}

	var req InteractRequest
	if err := json.Unmarshal(data, &req); err != nil {
		c.sendError("invalid_request", "Неверный формат запроса взаимодействия")
		return
	}

	// Передаем взаимодействие в игру
	result, err := c.Server.Game.HandleInteract(c.Info.PlayerID, req.ObjectID, req.InteractionIdx)
	if err != nil {
		c.sendError("interact_failed", err.Error())
		return
	}

	// Отправляем результат
	msg := Message{
		Type:    MsgInteractionResult,
		Payload: result,
		Time:    Now(),
		Seq:     c.getNextSeq(),
	}

	c.sendMessage(msg)
}

// sendMessage отправляет структурированное сообщение
func (c *Client) sendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Ошибка маршалинга сообщения: %v", err)
		return
	}

	c.sendRaw(data)
}

// sendRaw отправляет сырые данные
func (c *Client) sendRaw(data []byte) {
	select {
	case c.Send <- data:
	default:
		log.Printf("Канал отправки клиента %s переполнен", c.Info.ID)
	}
}

// sendError отправляет сообщение об ошибке
func (c *Client) sendError(code, message string) {
	msg := Message{
		Type: MsgError,
		Payload: ErrorMessage{
			Code:    code,
			Message: message,
		},
		Time: Now(),
		Seq:  c.getNextSeq(),
	}

	c.sendMessage(msg)
	log.Printf("Ошибка для клиента %s: %s - %s", c.Info.ID, code, message)
}

// sendWelcome отправляет приветственное сообщение
func (c *Client) sendWelcome() {
	msg := Message{
		Type: MsgCharacterUpdate,
		Payload: map[string]interface{}{
			"message":     "Добро пожаловать на сервер LOIL",
			"version":     "1.0.0",
			"server_time": Now(),
		},
		Time: Now(),
		Seq:  c.getNextSeq(),
	}

	c.sendMessage(msg)
}

func (c *Client) getNextSeq() int64 {
	c.sequence++
	return c.sequence
}
