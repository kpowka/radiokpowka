// Purpose: Runtime configuration from environment variables.

package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port string

	// DB
	DBDialect string // postgres|mysql|sqlite
	DBDSN     string

	// Security
	JWTSecret string

	// CORS
	CORSOrigins []string

	// Admin seed
	AdminUsername string
	AdminPassword string

	// Webhook shared secret
	WebhookSecret string

	// Tools
	YTDLPPath  string
	FFMPEGPath string

	// Twitch bot (optional)
	RunTwitchBot          bool
	TwitchNick            string
	TwitchOAuthToken      string
	TwitchChannel         string
	TwitchSpamEnabled     bool
	TwitchSpamMax         int
	TwitchSpamDelayMs     int
	TwitchRateLimitPerMin int
}

func MustLoad() Config {
	port := getEnv("PORT", "8080")

	dbDialect := getEnv("DB_DIALECT", "sqlite")
	dbDSN := getEnv("DB_DSN", "file:radiokpowka.db?_foreign_keys=on")

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	originsRaw := getEnv("CORS_ORIGINS", "http://localhost:5173")
	origins := splitCSV(originsRaw)

	adminUser := getEnv("ADMIN_USERNAME", "admin")
	adminPass := getEnv("ADMIN_PASSWORD", "admin123")

	webhookSecret := getEnv("WEBHOOK_SECRET", "")

	ytDlp := getEnv("YTDLP_PATH", "yt-dlp")
	ffmpeg := getEnv("FFMPEG_PATH", "ffmpeg")

	runBot := getEnvBool("RUN_TWITCH_BOT", false)
	tNick := getEnv("TWITCH_NICK", "")
	tTok := getEnv("TWITCH_OAUTH_TOKEN", "")
	tChan := getEnv("TWITCH_CHANNEL", "")
	spamEnabled := getEnvBool("TWITCH_SPAM_ENABLED", true)
	spamMax := getEnvInt("TWITCH_SPAM_MAX", 5)
	spamDelay := getEnvInt("TWITCH_SPAM_DELAY_MS", 900)
	rateLimit := getEnvInt("TWITCH_GLOBAL_RATE_LIMIT_PER_MIN", 18)

	return Config{
		Port: port,

		DBDialect: dbDialect,
		DBDSN:     dbDSN,

		JWTSecret: jwtSecret,

		CORSOrigins: origins,

		AdminUsername: adminUser,
		AdminPassword: adminPass,

		WebhookSecret: webhookSecret,

		YTDLPPath:  ytDlp,
		FFMPEGPath: ffmpeg,

		RunTwitchBot:          runBot,
		TwitchNick:            tNick,
		TwitchOAuthToken:      tTok,
		TwitchChannel:         tChan,
		TwitchSpamEnabled:     spamEnabled,
		TwitchSpamMax:         spamMax,
		TwitchSpamDelayMs:     spamDelay,
		TwitchRateLimitPerMin: rateLimit,
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnvBool(k string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	v = strings.ToLower(v)
	return v == "1" || v == "true" || v == "yes" || v == "y"
}

func getEnvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
