// Purpose: Command parsing + rate limiter helpers.

package bot

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

type CommandKind int

const (
	CmdNone CommandKind = iota
	CmdTrack
	CmdTrackSpam
)

type Command struct {
	Kind CommandKind
	N    int
}

func ParseCommand(text string) Command {
	t := strings.TrimSpace(text)
	if t == "" {
		return Command{Kind: CmdNone}
	}

	parts := strings.Fields(t)
	if len(parts) == 0 {
		return Command{Kind: CmdNone}
	}

	if parts[0] != "!track" {
		return Command{Kind: CmdNone}
	}
	if len(parts) == 1 {
		return Command{Kind: CmdTrack}
	}
	// !track spam N
	if len(parts) == 3 && parts[1] == "spam" {
		n, err := strconv.Atoi(parts[2])
		if err != nil {
			return Command{Kind: CmdTrackSpam, N: -1}
		}
		return Command{Kind: CmdTrackSpam, N: n}
	}

	return Command{Kind: CmdTrack}
}

// RateLimiter: token bucket based on "max per window"
type RateLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	tokens int
	reset  time.Time
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		max:    max,
		window: window,
		tokens: max,
		reset:  time.Now().Add(window),
	}
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.After(r.reset) {
		r.tokens = r.max
		r.reset = now.Add(r.window)
	}
	if r.tokens <= 0 {
		return false
	}
	r.tokens--
	return true
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
