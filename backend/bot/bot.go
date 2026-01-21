// Purpose: Twitch IRC bot runner.

package bot

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Nick       string
	OAuthToken string
	Channel    string

	SpamEnabled bool
	SpamMax     int
	SpamDelay   time.Duration

	GlobalRateLimitPerMin int

	// Returns a text to post for !track command
	GetCurrentTrackText func() string
}

func Run(cfg Config) error {
	if strings.TrimSpace(cfg.Nick) == "" ||
		strings.TrimSpace(cfg.OAuthToken) == "" ||
		strings.TrimSpace(cfg.Channel) == "" {
		return errors.New("twitch bot config incomplete: TWITCH_NICK/TWITCH_OAUTH_TOKEN/TWITCH_CHANNEL required")
	}
	if cfg.SpamMax <= 0 {
		cfg.SpamMax = 5
	}
	if cfg.SpamDelay <= 0 {
		cfg.SpamDelay = 900 * time.Millisecond
	}
	if cfg.GlobalRateLimitPerMin <= 0 {
		cfg.GlobalRateLimitPerMin = 18
	}
	if cfg.GetCurrentTrackText == nil {
		cfg.GetCurrentTrackText = func() string { return "RadioKpowka: трек неизвестен." }
	}

	conn, err := DialIRC(cfg.Nick, cfg.OAuthToken)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Join(cfg.Channel); err != nil {
		return err
	}

	log.Printf("twitch bot connected: #%s as %s", cfg.Channel, cfg.Nick)

	lim := NewRateLimiter(cfg.GlobalRateLimitPerMin, time.Minute)

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		cmd := ParseCommand(msg.Text)
		if cmd.Kind == CmdNone {
			continue
		}

		switch cmd.Kind {
		case CmdTrack:
			if !lim.Allow() {
				continue
			}
			_ = conn.Say(cfg.Channel, cfg.GetCurrentTrackText())

		case CmdTrackSpam:
			if !cfg.SpamEnabled {
				if lim.Allow() {
					_ = conn.Say(cfg.Channel, "Спам-команды отключены.")
				}
				continue
			}
			n := cmd.N
			if n <= 0 || n > cfg.SpamMax {
				if lim.Allow() {
					_ = conn.Say(cfg.Channel, "Неверное число. Максимум: "+itoa(cfg.SpamMax))
				}
				continue
			}

			// run spam in goroutine so bot continues reading messages
			wg.Add(1)
			go func() {
				defer wg.Done()
				text := cfg.GetCurrentTrackText()
				for i := 0; i < n; i++ {
					if !lim.Allow() {
						time.Sleep(cfg.SpamDelay)
						continue
					}
					_ = conn.Say(cfg.Channel, text)
					time.Sleep(cfg.SpamDelay)
				}
			}()
		}
	}
}
