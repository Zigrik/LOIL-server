package db

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

var dbLocations *sql.DB

const schemaLocations string = `
CREATE TABLE IF NOT EXISTS locations (
    id SERIAL PRIMARY KEY,
    level INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    location_type VARCHAR(50) NOT NULL,
    
    -- Перекрестки (ссылки на другие локации)
    crossroad_left1 INTEGER DEFAULT 0,
    crossroad_left2 INTEGER DEFAULT 0,
    crossroad_right1 INTEGER DEFAULT 0,
    crossroad_right2 INTEGER DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_locations_level ON locations(level);
CREATE INDEX IF NOT EXISTS idx_locations_type ON locations(location_type);;`

type Location struct {
	ID           int    `json:"id"`
	Level        int    `json:"level"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	LocationType string `json:"location_type"`

	// Перекрестки
	CrossroadLeft1  int `json:"crossroad_left1"`
	CrossroadLeft2  int `json:"crossroad_left2"`
	CrossroadRight1 int `json:"crossroad_right1"`
	CrossroadRight2 int `json:"crossroad_right2"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
