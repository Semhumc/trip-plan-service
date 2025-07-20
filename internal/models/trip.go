package models

import "time"

type Trip struct {
	ID               int32      `json:"id"`
	UserID           string     `json:"user_id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	StartPoint       Location   `json:"start_point"`
	EndPoint         Location   `json:"end_point"`
	Locations        []Location `json:"locations"`
	Nights           int        `json:"nights"`
	SingleOrMultiple bool       `json:"single_or_multiple"`
	CampOrCaravan    bool       `json:"camp_or_caravan"`
	StartDate        time.Time  `json:"start_date"`
	EndDate          time.Time  `json:"end_date"`
	MaxParticipants  int        `json:"max_participants"`
	Budget           float64    `json:"budget"`
	CreatedBy        int32      `json:"created_by"`
	LastModifiedBy   int32      `json:"last_modified_by"`
	IsFavorite       bool       `json:"is_favorite"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Location struct {
	ID           int32   `json:"id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Address      string  `json:"address"`
	Description  string  `json:"description"`
	ImageURL     string  `json:"image_url"`
	ContactInfo  string  `json:"contact_info"`
	Facilities   string  `json:"facilities"`
	OpeningHours string  `json:"opening_hours"`
}
