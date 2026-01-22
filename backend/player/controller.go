// Purpose: Server-authoritative playback state, queue transitions, WS broadcasts.
// Note: Audio is streamed per-client via /stream, but server state controls pause/play and position.

package player

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radiokpowka/backend/websocket"
	"radiokpowka/backend/youtube"
)

type ControllerDeps struct {
	DB  *gorm.DB
	Hub *websocket.Hub
	YT  *youtube.Client
}

type Controller struct {
	db  *gorm.DB
	hub *websocket.Hub
	yt  *youtube.Client

	mu sync.RWMutex
	rt runtime
}

func NewController(d ControllerDeps) *Controller {
	c := &Controller{
		db:  d.DB,
		hub: d.Hub,
		yt:  d.YT,
	}
	// defaults
	c.rt.volume = 0.8
	c.rt.isPaused = true
	c.rt.isPlaying = false

	// background auto-advance based on duration (approx)
	go c.autoAdvanceLoop()

	// initialize current from DB if exists
	_ = c.refreshCurrentFromDB()
	c.autoStartIfStopped()

	// broadcast initial state
	c.broadcastState()
	c.broadcastQueue()

	return c
}

func (c *Controller) State() PlayerState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pos := c.positionLocked()
	st := PlayerState{
		IsPlaying:   c.rt.isPlaying,
		IsPaused:    c.rt.isPaused,
		Volume:      c.rt.volume,
		PositionSec: pos,
		DurationSec: c.rt.durationSec,
	}
	if c.rt.currentTrackID != "" {
		st.Current = &TrackDTO{
			ID:          c.rt.currentTrackID,
			Title:       c.rt.currentTitle,
			URL:         c.rt.currentURL,
			AddedByNick: c.rt.currentAddedBy,
		}
	}
	return st
}

func (c *Controller) Play() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.refreshCurrentFromDBLocked(); err != nil {
		return err
	}
	if c.rt.currentTrackID == "" {
		return errors.New("queue is empty")
	}

	if !c.rt.isPlaying {
		c.rt.isPlaying = true
	}
	if c.rt.isPaused {
		c.rt.isPaused = false
		c.rt.startedAt = time.Now().UTC()
		// basePosSec remains as stored
	}
	c.broadcastStateLocked()
	return nil
}

func (c *Controller) Pause() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.rt.isPlaying {
		// still broadcast paused state
		c.rt.isPaused = true
		c.broadcastStateLocked()
		return nil
	}

	// capture position
	c.rt.basePosSec = c.positionLocked()
	c.rt.isPaused = true
	c.rt.startedAt = time.Time{} // reset
	c.broadcastStateLocked()
	return nil
}

func (c *Controller) Next() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	tx := c.db.Begin()
	defer func() { _ = tx.Rollback() }()

	q, t, err := nextTrack(tx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// end of queue: stop
			c.rt.isPlaying = false
			c.rt.isPaused = true
			c.rt.basePosSec = 0
			c.rt.startedAt = time.Time{}
			c.rt.currentQueueID = ""
			c.rt.currentTrackID = ""
			c.rt.currentTitle = ""
			c.rt.currentURL = ""
			c.rt.currentAddedBy = ""
			c.rt.durationSec = 0
			_ = tx.Commit()
			c.broadcastStateLocked()
			c.broadcastQueueLocked()
			return nil
		}
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}

	c.applyCurrentLocked(q.ID.String(), t.ID.String(), t.Title, t.SourceURL, t.AddedByNick, t.DurationSec)
	c.broadcastQueueLocked()
	c.broadcastStateLocked()
	return nil
}

func (c *Controller) Prev() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	tx := c.db.Begin()
	defer func() { _ = tx.Rollback() }()

	q, t, err := prevTrack(tx)
	if err != nil {
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}

	c.applyCurrentLocked(q.ID.String(), t.ID.String(), t.Title, t.SourceURL, t.AddedByNick, t.DurationSec)
	c.broadcastQueueLocked()
	c.broadcastStateLocked()
	return nil
}

func (c *Controller) SetVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	c.mu.Lock()
	c.rt.volume = v
	c.broadcastStateLocked()
	c.mu.Unlock()
}

func (c *Controller) AddTrack(url string, addedByUser *uuid.UUID, addedByNick string, insertNext bool, isDonation bool) (string, error) {
	// Resolve meta(s)
	metas, err := c.yt.ResolveMetas(context.Background(), url)
	if err != nil {
		return "", err
	}
	if len(metas) == 0 {
		return "", errors.New("no tracks resolved")
	}

	tx := c.db.Begin()
	defer func() { _ = tx.Rollback() }()

	// Determine insert position
	pos, err := maxPosition(tx)
	if err != nil {
		return "", err
	}

	status := "next"
	insertPos := pos + 1

	// InsertNext means: position right after current
	if insertNext {
		// find current position
		var cur struct{ Position int }
		if err := tx.Table("queue_entries").Select("position").Where("status = ?", "current").Scan(&cur).Error; err == nil && cur.Position > 0 {
			insertPos = cur.Position + 1
			if err := shiftPositionsFrom(tx, insertPos, len(metas)); err != nil {
				return "", err
			}
		}
	}

	inserted := make([]struct {
		queueID string
		track   TrackDTO
	}, 0, len(metas))
	for i, meta := range metas {
		t, err := addTrack(tx, meta.WebpageURL, meta.Title, meta.DurationSec, addedByUser, addedByNick)
		if err != nil {
			return "", err
		}

		q, err := insertQueueEntry(tx, t.ID, insertPos+i, status, isDonation)
		if err != nil {
			return "", err
		}
		inserted = append(inserted, struct {
			queueID string
			track   TrackDTO
		}{
			queueID: q.ID.String(),
			track: TrackDTO{
				ID:          t.ID.String(),
				Title:       t.Title,
				URL:         t.SourceURL,
				AddedByNick: t.AddedByNick,
			},
		})
	}

	if err := ensureQueueHasCurrent(tx); err != nil {
		return "", err
	}

	if err := tx.Commit().Error; err != nil {
		return "", err
	}

	// refresh current runtime if empty
	_ = c.refreshCurrentFromDB()
	c.autoStartIfStopped()

	// broadcast
	for _, item := range inserted {
		c.hub.Broadcast(websocket.Event{Type: websocket.EventTrackAdded, Data: map[string]any{
			"id":          item.queueID,
			"title":       item.track.Title,
			"url":         item.track.URL,
			"addedByNick": item.track.AddedByNick,
			"isDonation":  isDonation,
		}})
	}
	c.broadcastQueue()
	c.broadcastState()

	return inserted[0].queueID, nil
}

func (c *Controller) ListQueue() ([]QueueEntryDTO, error) {
	return listQueue(c.db)
}

func (c *Controller) CurrentForStreaming() (url string, posSec int, vol float64, paused bool, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.rt.currentURL == "" {
		return "", 0, c.rt.volume, true, false
	}
	return c.rt.currentURL, c.positionLocked(), c.rt.volume, c.rt.isPaused || !c.rt.isPlaying, true
}

func (c *Controller) refreshCurrentFromDB() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.refreshCurrentFromDBLocked()
}

func (c *Controller) refreshCurrentFromDBLocked() error {
	curQ, curT, err := getCurrent(c.db)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// empty
			c.applyCurrentLocked("", "", "", "", "", 0)
			return nil
		}
		return err
	}
	c.applyCurrentLocked(curQ.ID.String(), curT.ID.String(), curT.Title, curT.SourceURL, curT.AddedByNick, curT.DurationSec)
	return nil
}

func (c *Controller) applyCurrentLocked(qid, tid, title, url, addedBy string, duration int) {
	c.rt.currentQueueID = qid
	c.rt.currentTrackID = tid
	c.rt.currentTitle = title
	c.rt.currentURL = url
	c.rt.currentAddedBy = addedBy
	c.rt.durationSec = duration

	// reset position when switching
	c.rt.basePosSec = 0
	if c.rt.isPlaying && !c.rt.isPaused {
		c.rt.startedAt = time.Now().UTC()
	} else {
		c.rt.startedAt = time.Time{}
	}
}

func (c *Controller) autoStartIfStopped() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.rt.currentTrackID == "" {
		return
	}
	if c.rt.isPlaying {
		return
	}

	c.rt.isPlaying = true
	c.rt.isPaused = false
	c.rt.basePosSec = 0
	c.rt.startedAt = time.Now().UTC()
}

func (c *Controller) positionLocked() int {
	if !c.rt.isPlaying || c.rt.isPaused || c.rt.startedAt.IsZero() {
		return c.rt.basePosSec
	}
	elapsed := int(time.Since(c.rt.startedAt).Seconds())
	p := c.rt.basePosSec + elapsed
	if c.rt.durationSec > 0 && p > c.rt.durationSec {
		p = c.rt.durationSec
	}
	return p
}

func (c *Controller) broadcastState() {
	c.mu.RLock()
	st := c.State()
	c.mu.RUnlock()
	c.hub.Broadcast(websocket.Event{Type: websocket.EventPlayerState, Data: st})
}

func (c *Controller) broadcastStateLocked() {
	st := c.State()
	c.hub.Broadcast(websocket.Event{Type: websocket.EventPlayerState, Data: st})
}

func (c *Controller) broadcastQueue() {
	items, err := c.ListQueue()
	if err != nil {
		return
	}
	c.hub.Broadcast(websocket.Event{Type: websocket.EventQueueUpdate, Data: items})
}

func (c *Controller) broadcastQueueLocked() {
	items, err := c.ListQueue()
	if err != nil {
		return
	}
	c.hub.Broadcast(websocket.Event{Type: websocket.EventQueueUpdate, Data: items})
}

func (c *Controller) autoAdvanceLoop() {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	for range t.C {
		c.mu.RLock()
		playing := c.rt.isPlaying && !c.rt.isPaused
		dur := c.rt.durationSec
		pos := c.positionLocked()
		c.mu.RUnlock()

		if playing && dur > 0 && pos >= dur {
			_ = c.Next()
		}
	}
}
