package models

import "time"

const (
	// DefaultUserID represents the legacy single-user watchlist owner.
	DefaultUserID = "default"
	// DefaultUserName is used when creating the initial profile.
	DefaultUserName = "Primary Profile"
)

// User models a NovaStream profile capable of holding watchlist data.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
