package db

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var dbUsers *sql.DB

const schemaUsers string = `
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    login TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    salt TEXT NOT NULL,

    game_nickname TEXT NOT NULL,
    family_name TEXT NOT NULL,

    last_ip TEXT,
    registration_ip TEXT,

    is_online INTEGER DEFAULT 0,
    is_banned INTEGER DEFAULT 0,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME,
    last_activity DATETIME
);`

type User struct {
	ID           int64  `json:"id"`
	Login        string `json:"login"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Salt         string `json:"-"`

	GameNickname string `json:"game_nickname"`
	FamilyName   string `json:"family_name"`

	LastIP         string `json:"last_ip"`
	RegistrationIP string `json:"registration_ip"`

	IsOnline bool `json:"is_online"`
	IsBanned bool `json:"is_banned"`

	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
	LastActivity time.Time `json:"last_activity"`
}

func CloseDatabase() {
	dbUsers.Close()
}

func Init(logger *log.Logger) error {

	dbUsersPath := filepath.Join("db", "users.db")

	var err error
	dbUsers, err = sql.Open("sqlite", dbUsersPath)
	if err != nil {
		return err
	}

	var install bool
	if _, err := os.Stat(dbUsersPath); os.IsNotExist(err) {
		install = true
	}

	if install {
		file, err := os.Create(dbUsersPath)
		if err != nil {
			return err
		}
		file.Close()
		logger.Printf("INFO: the %s file has been created\n", dbUsersPath)

		_, err = dbUsers.Exec(schemaUsers)
		if err != nil {
			return err
		}
	}

	logger.Printf("INFO: the %s database is ready for use\n", dbUsersPath)
	return nil
}

// Генерация соли
func generateSalt() string {
	return fmt.Sprintf("%x", rand.Int63())
}

// Хеширование пароля (простой пример)
func hashPassword(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(password + salt))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Проверка пароля
func checkPassword(password, hash, salt string) bool {
	return hashPassword(password, salt) == hash
}

// Регистрация пользователя
func CreateUser(login, email, password, nickname, familyName, ip string) error {
	salt := generateSalt()
	hash := hashPassword(password, salt)

	_, err := dbUsers.Exec(`
        INSERT INTO users 
        (login, email, password_hash, salt, game_nickname, family_name, registration_ip, last_ip)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `, login, email, hash, salt, nickname, familyName, ip, ip)

	return err
}

// Авторизация пользователя
func AuthenticateUser(login, password, ip string) (*User, error) {
	var user User
	err := dbUsers.QueryRow(`
        SELECT id, login, email, password_hash, salt, game_nickname, family_name, 
               last_ip, is_online, is_banned, created_at, last_login
        FROM users 
        WHERE login = ? OR email = ?
    `, login, login).Scan(
		&user.ID, &user.Login, &user.Email, &user.PasswordHash, &user.Salt,
		&user.GameNickname, &user.FamilyName, &user.LastIP,
		&user.IsOnline, &user.IsBanned, &user.CreatedAt, &user.LastLogin,
	)

	if err != nil {
		return nil, err
	}

	// Проверка пароля
	if !checkPassword(password, user.PasswordHash, user.Salt) {
		return nil, errors.New("invalid password")
	}

	// Обновляем IP и время входа
	_, err = dbUsers.Exec(`
        UPDATE users 
        SET last_ip = ?, last_login = CURRENT_TIMESTAMP, is_online = 1 
        WHERE id = ?
    `, ip, user.ID)

	if err != nil {
		return nil, err
	}

	user.LastIP = ip
	user.IsOnline = true

	return &user, nil
}

// Выход пользователя
func LogoutUser(dbUsers *sql.DB, userID int64) error {
	_, err := dbUsers.Exec(`
        UPDATE users 
        SET is_online = 0, last_activity = CURRENT_TIMESTAMP 
        WHERE id = ?
    `, userID)
	return err
}
