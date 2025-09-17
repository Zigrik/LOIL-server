package db

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

var dbLocations *sql.DB

const schemaLocations string = `
CREATE TABLE IF NOT EXISTS locations (
    id SERIAL PRIMARY KEY,
    level INTEGER NOT NULL DEFAULT 1,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    location_type VARCHAR(50) NOT NULL DEFAULT 'test',
    
    crossroad_north INTEGER DEFAULT 0,
    crossroad_south INTEGER DEFAULT 0,
    crossroad_east INTEGER DEFAULT 0,
    crossroad_west INTEGER DEFAULT 0,
    
    terrain INTEGER[] DEFAULT '{}',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_locations_level ON locations(level);
CREATE INDEX IF NOT EXISTS idx_locations_terrain ON locations USING GIN (terrain);`

type Location struct {
	ID           int    `json:"id"`
	Level        int    `json:"level"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	LocationType string `json:"location_type"`

	// Отдельные перекрестки (переходы)
	CrossroadNorth int `json:"crossroad_north"`
	CrossroadSouth int `json:"crossroad_south"`
	CrossroadEast  int `json:"crossroad_east"`
	CrossroadWest  int `json:"crossroad_west"`

	// Массив земли (тайлы)
	Terrain []int `json:"terrain"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
