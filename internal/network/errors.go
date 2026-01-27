package network

import "errors"

// GameError - ошибка игры для сети
type GameError struct {
	Code    string
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}

// NewError создает новую ошибку
func NewError(code, message string) error {
	return &GameError{
		Code:    code,
		Message: message,
	}
}

// IsGameError проверяет, является ли ошибка GameError
func IsGameError(err error) bool {
	var gameErr *GameError
	return errors.As(err, &gameErr)
}

// GetErrorCode возвращает код ошибки
func GetErrorCode(err error) string {
	var gameErr *GameError
	if errors.As(err, &gameErr) {
		return gameErr.Code
	}
	return "internal_error"
}
