// Purpose: DB operations for queue/tracks (portable).

package player

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radiokpowka/backend/db"
)

func ensureQueueHasCurrent(tx *gorm.DB) error {
	var cur db.QueueEntry
	err := tx.Where("status = ?", "current").First(&cur).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// set first next as current if exists
	var first db.QueueEntry
	if err := tx.Where("status = ?", "next").Order("position asc").First(&first).Error; err != nil {
		return nil // empty queue is OK
	}
	return tx.Model(&db.QueueEntry{}).Where("id = ?", first.ID).Update("status", "current").Error
}

func listQueue(tx *gorm.DB) ([]QueueEntryDTO, error) {
	// join queue_entries + tracks
	type row struct {
		QID        string
		Status     string
		Position   int
		AddedAt    time.Time
		IsDonation bool
		Title      string
		URL        string
		AddedByNick string
	}
	var rows []row
	err := tx.Table("queue_entries").
		Select("queue_entries.id as qid, queue_entries.status, queue_entries.position, queue_entries.added_at, queue_entries.is_donation, tracks.title, tracks.source_url as url, tracks.added_by_nick").
		Joins("join tracks on tracks.id = queue_entries.track_id").
		Order("queue_entries.position asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]QueueEntryDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, QueueEntryDTO{
			ID:          r.QID,
			Title:       r.Title,
			URL:         r.URL,
			AddedByNick: r.AddedByNick,
			AddedAt:     r.AddedAt.UTC().Format(time.RFC3339),
			Status:      r.Status,
			IsDonation:  r.IsDonation,
		})
	}
	return out, nil
}

func addTrack(tx *gorm.DB, url, title string, duration int, addedByUser *uuid.UUID, addedByNick string) (db.Track, error) {
	t := db.Track{
		ID:            uuid.New(),
		Title:         title,
		SourceURL:     url,
		DurationSec:   duration,
		AddedByUserID: addedByUser,
		AddedByNick:   addedByNick,
		MetadataJSON:  []byte(`{}`),
		CreatedAt:     time.Now().UTC(),
	}
	if err := tx.Create(&t).Error; err != nil {
		return db.Track{}, err
	}
	return t, nil
}

func maxPosition(tx *gorm.DB) (int, error) {
	var max int
	if err := tx.Model(&db.QueueEntry{}).Select("COALESCE(MAX(position), 0)").Scan(&max).Error; err != nil {
		return 0, err
	}
	return max, nil
}

func insertQueueEntry(tx *gorm.DB, trackID uuid.UUID, pos int, status string, isDonation bool) (db.QueueEntry, error) {
	q := db.QueueEntry{
		ID:         uuid.New(),
		TrackID:    trackID,
		Position:   pos,
		Status:     status,
		AddedAt:    time.Now().UTC(),
		IsDonation: isDonation,
	}
	if err := tx.Create(&q).Error; err != nil {
		return db.QueueEntry{}, err
	}
	return q, nil
}

func shiftPositionsFrom(tx *gorm.DB, fromPos int) error {
	// shift all >= fromPos by +1
	return tx.Model(&db.QueueEntry{}).
		Where("position >= ?", fromPos).
		Update("position", gorm.Expr("position + 1")).Error
}

func nextTrack(tx *gorm.DB) (db.QueueEntry, db.Track, error) {
	var cur db.QueueEntry
	if err := tx.Where("status = ?", "current").First(&cur).Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}

	// current -> prev
	if err := tx.Model(&db.QueueEntry{}).Where("id = ?", cur.ID).Update("status", "prev").Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}

	// pick first next
	var nx db.QueueEntry
	if err := tx.Where("status = ?", "next").Order("position asc").First(&nx).Error; err != nil {
		// no next: stop playback (queue ended)
		return db.QueueEntry{}, db.Track{}, gorm.ErrRecordNotFound
	}

	if err := tx.Model(&db.QueueEntry{}).Where("id = ?", nx.ID).Update("status", "current").Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}

	var t db.Track
	if err := tx.Where("id = ?", nx.TrackID).First(&t).Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}
	return nx, t, nil
}

func prevTrack(tx *gorm.DB) (db.QueueEntry, db.Track, error) {
	// current -> next, last prev -> current
	var cur db.QueueEntry
	if err := tx.Where("status = ?", "current").First(&cur).Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}

	var pv db.QueueEntry
	if err := tx.Where("status = ?", "prev").Order("position desc").First(&pv).Error; err != nil {
		return db.QueueEntry{}, db.Track{}, errors.New("no prev")
	}

	if err := tx.Model(&db.QueueEntry{}).Where("id = ?", cur.ID).Update("status", "next").Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}
	if err := tx.Model(&db.QueueEntry{}).Where("id = ?", pv.ID).Update("status", "current").Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}

	var t db.Track
	if err := tx.Where("id = ?", pv.TrackID).First(&t).Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}
	return pv, t, nil
}

func getCurrent(tx *gorm.DB) (db.QueueEntry, db.Track, error) {
	if err := ensureQueueHasCurrent(tx); err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}
	var cur db.QueueEntry
	if err := tx.Where("status = ?", "current").First(&cur).Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}
	var t db.Track
	if err := tx.Where("id = ?", cur.TrackID).First(&t).Error; err != nil {
		return db.QueueEntry{}, db.Track{}, err
	}
	return cur, t, nil
}
