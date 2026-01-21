// Purpose: Entry point. Loads config, connects DB, seeds admin user, starts Gin HTTP server.
// Also optionally starts Twitch bot in the same process if RUN_TWITCH_BOT=true.

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"radiokpowka/backend/api"
	"radiokpowka/backend/bot"
	"radiokpowka/backend/config"
	"radiokpowka/backend/db"
)

func main() {
	cfg := config.MustLoad()

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	if err := db.SeedAdmin(database, cfg); err != nil {
		log.Fatalf("seed admin failed: %v", err)
	}

	handler := api.NewRouter(cfg, database)

	// Optionally start Twitch bot (non-blocking)
	if cfg.RunTwitchBot {
		go func() {
			if err := bot.Run(bot.Config{
				Nick:                  cfg.TwitchNick,
				OAuthToken:            cfg.TwitchOAuthToken,
				Channel:               cfg.TwitchChannel,
				SpamEnabled:           cfg.TwitchSpamEnabled,
				SpamMax:               cfg.TwitchSpamMax,
				SpamDelay:             time.Duration(cfg.TwitchSpamDelayMs) * time.Millisecond,
				GlobalRateLimitPerMin: cfg.TwitchRateLimitPerMin,
				GetCurrentTrackText: func() string {
					// The bot will call backend logic via HTTP in a future iteration.
					// For now, it prints a placeholder. In "tests later" mode, keep it simple.
					// You can later wire it to an internal player controller getter.
					return "RadioKpowka: трек пока недоступен (заглушка)."
				},
			}); err != nil {
				log.Printf("twitch bot stopped: %v", err)
			}
		}()
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("RadioKpowka backend listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("shutdown: starting...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("shutdown: done")
}
