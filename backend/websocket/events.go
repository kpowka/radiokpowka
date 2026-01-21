// Purpose: Event envelope types for realtime.

package websocket

type EventType string

const (
	EventPlayerState     EventType = "player_state"
	EventQueueUpdate     EventType = "queue_update"
	EventTrackAdded      EventType = "track_added"
	EventDonationReceived EventType = "donation_received"
	EventAuthUpdate      EventType = "auth_update"
)

type Event struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}
