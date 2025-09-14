package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

const dbUsers string = "/db/users.db"

const schema string = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
	title VARCHAR NOT NULL DEFAULT "",
	comment TEXT,
	repeat VARCHAR(128)
);`

var database *sql.DB

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
	database.Close()
}

func Init(logger *log.Logger) error {

	var err error
	database, err = sql.Open("sqlite", dbUsers)
	if err != nil {
		return err
	}

	var install bool
	if _, err := os.Stat(dbUsers); os.IsNotExist(err) {
		install = true
	}

	if install {
		file, err := os.Create(dbUsers)
		if err != nil {
			return err
		}
		file.Close()
		logger.Printf("INFO: the %s file has been created\n", dbUsers)

		_, err = database.Exec(schema)
		if err != nil {
			return err
		}
	}

	logger.Printf("INFO: the %s database is ready for use\n", dbUsers)
	return nil
}
