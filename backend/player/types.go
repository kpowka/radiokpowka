// Purpose: Player state and DTOs.

package player

import "time"

type TrackDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	AddedByNick string `json:"addedByNick,omitempty"`
}

type PlayerState struct {
	IsPlaying   bool     `json:"isPlaying"`
	IsPaused    bool     `json:"isPaused"`
	Volume      float64  `json:"volume"`
	PositionSec int      `json:"positionSec"`
	DurationSec int      `json:"durationSec"`
	Current     *TrackDTO `json:"current,omitempty"`
}

type QueueEntryDTO struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	AddedByNick string `json:"addedByNick"`
	AddedAt    string `json:"addedAt"`
	Status     string `json:"status"` // prev|current|next
	IsDonation bool   `json:"isDonation,omitempty"`
}

type runtime struct {
	isPlaying   bool
	isPaused    bool
	volume      float64

	currentQueueID string
	currentTrackID string
	currentTitle   string
	currentURL     string
	currentAddedBy string
	durationSec    int

	startedAt     time.Time
	basePosSec    int // position at startedAt
}
