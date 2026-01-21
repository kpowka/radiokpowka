// Purpose: Core models (portable between MySQL/SQLite/Postgres).
// Note: Migrations are provided as SQL in /migrations/* (Part 3).
// GORM models are used for runtime queries.

package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type User struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:32;not null" json:"role"` // owner|listener
	CreatedAt    time.Time `json:"created_at"`
}

type Track struct {
	ID            uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Title         string         `gorm:"size:512;not null" json:"title"`
	SourceURL     string         `gorm:"size:2048;not null" json:"source_url"`
	DurationSec   int            `gorm:"not null;default:0" json:"duration"`
	AddedByUserID *uuid.UUID     `gorm:"type:char(36)" json:"added_by_user_id,omitempty"`
	AddedByNick   string         `gorm:"size:128" json:"added_by_nick"`
	MetadataJSON  datatypes.JSON `gorm:"type:json" json:"metadata_json"`
	CreatedAt     time.Time      `json:"created_at"`
}

type QueueEntry struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	TrackID   uuid.UUID `gorm:"type:char(36);not null;index" json:"track_id"`
	Position  int       `gorm:"not null;index" json:"position"`
	Status    string    `gorm:"size:16;not null;index" json:"status"` // prev|current|next
	AddedAt   time.Time `gorm:"not null" json:"added_at"`
	IsDonation bool     `gorm:"not null;default:false" json:"is_donation"`
}

type Donation struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	DonorNick string    `gorm:"size:128;not null" json:"donor_nick"`
	Amount    int64     `gorm:"not null;default:0" json:"amount"`
	Message   string    `gorm:"type:text" json:"message"`
	TrackURL  string    `gorm:"size:2048" json:"track_url"`
	CreatedAt time.Time `json:"created_at"`
}

type Integration struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Type        string         `gorm:"size:64;not null;index" json:"type"` // twitch|donationalerts|donx
	ConfigJSON  datatypes.JSON `gorm:"type:json" json:"config_json"`
	ConnectedAt *time.Time     `json:"connected_at,omitempty"`
}

type Playlist struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	URL       string    `gorm:"size:2048;not null" json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
