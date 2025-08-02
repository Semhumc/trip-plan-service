package models

import "time"

type Trip struct {
	ID            int       `json:"id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`
	StartPosition string    `json:"start_position"`
	EndPosition   string    `json:"end_position"`
	Description   string    `json:"description,omitempty"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Location struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Address   *string   `json:"address,omitempty"`
	SiteURL   *string   `json:"site_url,omitempty"`
	Latitude  float64   `json:"latitude"`   // enlem
    Longitude float64   `json:"longitude"`  // boylam
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type TripLocation struct {
	TripID     int `json:"trip_id"`
	LocationID int `json:"location_id"`
	Position   int `json:"position"`
}

type TripWithLocations struct {
	Trip      Trip       `json:"trip"`
	Locations []Location `json:"locations"`
}
