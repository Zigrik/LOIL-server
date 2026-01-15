package world

type Transition struct {
	LocationID int    `json:"location_id"`
	Type       string `json:"type"`
}

type Location struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	RoadTiles   []int                  `json:"road_tiles"`
	Transitions map[string]*Transition `json:"transitions"`
}

type Character struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Location   int     `json:"location"`
	X          float64 `json:"x"`
	Speed      float64 `json:"speed"`
	Direction  int     `json:"direction"`
	Controlled int     `json:"controlled"`
	Vertical   int     `json:"-"`
}

type World struct {
	PlayerID   int          `json:"player_id"`
	Characters []*Character `json:"characters"`
	Locations  []*Location  `json:"locations"`
}
