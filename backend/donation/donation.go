// Purpose: Donation webhook parsing + insert-next track logic.

package donations

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radiokpowka/backend/db"
	"radiokpowka/backend/player"
	"radiokpowka/backend/websocket"
)

type IncomingDonation struct {
	DonorNick string `json:"donor_nick"`
	Amount    int64  `json:"amount"`
	Message   string `json:"message"`
	// Optional explicit track_url, if provider sends it separately:
	TrackURL string `json:"track_url,omitempty"`
}

type Deps struct {
	DB     *gorm.DB
	Player *player.Controller
	Hub    *websocket.Hub
}

var ytURL = regexp.MustCompile(`https?://(www\.)?(youtube\.com/watch\?v=[\w-]+|youtu\.be/[\w-]+)`)
var ytPlaylist = regexp.MustCompile(`https?://(www\.)?youtube\.com/playlist\?list=[\w-]+`)

func HandleDonation(ctx context.Context, d Deps, payload IncomingDonation) error {
	if strings.TrimSpace(payload.DonorNick) == "" {
		payload.DonorNick = "donor"
	}

	msg := strings.TrimSpace(payload.Message)
	trackURL := strings.TrimSpace(payload.TrackURL)

	if trackURL == "" && msg != "" {
		trackURL = extractFirstTrackURL(msg)
	}

	// persist donation (always)
	row := db.Donation{
		ID:        uuid.New(),
		DonorNick: payload.DonorNick,
		Amount:    payload.Amount,
		Message:   payload.Message,
		TrackURL:  trackURL,
		CreatedAt: time.Now().UTC(),
	}
	_ = d.DB.Create(&row).Error

	d.Hub.Broadcast(websocket.Event{
		Type: websocket.EventDonationReceived,
		Data: map[string]any{
			"donorNick": payload.DonorNick,
			"amount":    payload.Amount,
			"message":   payload.Message,
			"trackUrl":  trackURL,
		},
	})

	// if we have track link -> insert next in queue
	if trackURL != "" {
		_, err := d.Player.AddTrack(trackURL, nil, payload.DonorNick, true, true)
		return err
	}

	return nil
}

func extractFirstTrackURL(msg string) string {
	if m := ytURL.FindString(msg); m != "" {
		return m
	}
	// if playlist link arrives, we still return it (backend can expand later)
	if m := ytPlaylist.FindString(msg); m != "" {
		return m
	}
	return ""
}

var _ = errors.New
