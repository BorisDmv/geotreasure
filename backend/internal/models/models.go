package models

import "time"

// User is a registered account. PasswordHash is never serialized to JSON.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Treasure is a buried treasure pinned to a real-world location.
type Treasure struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	OwnerName   string    `json:"ownerName"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Hint        string    `json:"hint"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	CreatedAt   time.Time `json:"createdAt"`

	// Populated when listing treasures.
	FoundByCount int  `json:"foundByCount"`
	FoundByMe    bool `json:"foundByMe"`
	// DistanceMeters is only set for "near" searches.
	DistanceMeters *float64 `json:"distanceMeters,omitempty"`
}
