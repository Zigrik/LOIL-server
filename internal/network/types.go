package network

import "time"

type GameStateProvider interface {
	// Получение состояния
	GetLocationState(locationID int) *LocationState
	GetCharactersInLocation(locationID int) []*CharacterState
	GetCreaturesInLocation(locationID int) []*CreatureState
	GetObjectsInLocation(locationID int) []*ObjectState
	GetCharacterByID(characterID int) *CharacterState

	// Обработка действий
	HandleJoin(playerID, characterID, locationID int) (*CharacterState, error)
	HandleMove(playerID int, direction, vertical int) error
	HandleStop(playerID int) error
	HandleInteract(playerID int, objectID, interactionIdx int) (*InteractionResult, error)

	// Утилиты
	GetServerTime() int64
	GetLocationName(locationID int) string
}

// MessageType - тип сообщения
type MessageType string

const (
	// От сервера к клиенту
	MsgWorldState        MessageType = "world_state"
	MsgLocationUpdate    MessageType = "location_update"
	MsgCharacterUpdate   MessageType = "character_update"
	MsgInteractionResult MessageType = "interaction_result"
	MsgError             MessageType = "error"
	MsgPing              MessageType = "ping"

	// От клиента к серверу
	MsgJoin     MessageType = "join"
	MsgMove     MessageType = "move"
	MsgStop     MessageType = "stop"
	MsgInteract MessageType = "interact"
	MsgPong     MessageType = "pong"
)

// Message - базовое сообщение
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Seq     int64       `json:"seq,omitempty"`  // Порядковый номер
	Time    int64       `json:"time,omitempty"` // Время отправки (unix ms)
}

// JoinRequest - запрос на присоединение
type JoinRequest struct {
	PlayerID    int    `json:"player_id"`
	CharacterID int    `json:"character_id,omitempty"`
	LocationID  int    `json:"location_id"`
	Token       string `json:"token,omitempty"` // Для будущей авторизации
}

// MoveRequest - запрос на движение
type MoveRequest struct {
	Direction int `json:"direction"` // -1: left, 0: stop, 1: right
	Vertical  int `json:"vertical"`  // -1: down, 0: none, 1: up
}

// InteractRequest - запрос на взаимодействие
type InteractRequest struct {
	ObjectID       int `json:"object_id"`
	InteractionIdx int `json:"interaction_idx"`
}

// CharacterState - состояние персонажа для сети
type CharacterState struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	LocationID int     `json:"location_id"`
	X          float64 `json:"x"`
	Direction  int     `json:"direction"`
	Speed      float64 `json:"speed"`
	Controlled int     `json:"controlled"`
	LastUpdate int64   `json:"last_update"`
}

// CreatureState - состояние существа для сети
type CreatureState struct {
	ID         int     `json:"id"`
	TypeID     int     `json:"type_id"`
	Name       string  `json:"name"`
	LocationID int     `json:"location_id"`
	X          float64 `json:"x"`
	Health     int     `json:"health"`
	MaxHealth  int     `json:"max_health"`
	Hunger     int     `json:"hunger,omitempty"`
	Behavior   string  `json:"behavior,omitempty"`
	LastUpdate int64   `json:"last_update"`
}

// ObjectState - состояние объекта для сети
type ObjectState struct {
	ID            int   `json:"id"`
	TypeID        int   `json:"type_id"`
	LocationID    int   `json:"location_id"`
	X             int   `json:"x"`
	Durability    int   `json:"durability"`
	MaxDurability int   `json:"max_durability"`
	GrowthStage   int   `json:"growth_stage,omitempty"`
	LastUpdate    int64 `json:"last_update"`
}

// LocationState - состояние локации для сети
type LocationState struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Width      int    `json:"width"`
	Foreground []int  `json:"foreground"`
	Road       []int  `json:"road"`
	Ground     []int  `json:"ground"`
	Background []int  `json:"background"`
	LastUpdate int64  `json:"last_update"`
}

// WorldState - полное состояние для клиента
type WorldState struct {
	PlayerID   int               `json:"player_id"`
	Location   *LocationState    `json:"location"`
	Characters []*CharacterState `json:"characters"`
	Creatures  []*CreatureState  `json:"creatures"`
	Objects    []*ObjectState    `json:"objects"`
	ServerTime int64             `json:"server_time"`
}

// LocationUpdate - обновление локации
type LocationUpdate struct {
	LocationID int               `json:"location_id"`
	Characters []*CharacterState `json:"characters,omitempty"`
	Creatures  []*CreatureState  `json:"creatures,omitempty"`
	Objects    []*ObjectState    `json:"objects,omitempty"`
	LayerHash  string            `json:"layer_hash,omitempty"` // Для будущих оптимизаций
	ServerTime int64             `json:"server_time"`
}

// CharacterUpdate - обновление персонажа
type CharacterUpdate struct {
	CharacterID int             `json:"character_id"`
	State       *CharacterState `json:"state,omitempty"`
	Removed     bool            `json:"removed,omitempty"`
	ServerTime  int64           `json:"server_time"`
}

// InteractionResult - результат взаимодействия
type InteractionResult struct {
	Success    bool            `json:"success"`
	ObjectID   int             `json:"object_id,omitempty"`
	Message    string          `json:"message,omitempty"`
	Items      []InventoryItem `json:"items,omitempty"`
	ServerTime int64           `json:"server_time"`
}

// InventoryItem - предмет инвентаря
type InventoryItem struct {
	ItemID int    `json:"item_id"`
	Count  int    `json:"count"`
	Name   string `json:"name,omitempty"`
}

// ErrorMessage - сообщение об ошибке
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Helper functions
func Now() int64 {
	return time.Now().UnixMilli()
}
