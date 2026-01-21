# RadioKpowka

Web music service with:
- Go backend (Gin) + WebSocket real-time updates
- React + TypeScript frontend (Vite)
- DB support: Postgres / MySQL / SQLite via GORM
- Server-authoritative playback (pause/resume on server)
- YouTube audio-only streaming via yt-dlp + ffmpeg (no video embed)
- Donation webhook: auto-insert track next if message contains a link
- Twitch bot: `!track` and `!track spam N` (rate-limited + configurable)

## Quick start (local)

### 1) Requirements
- Go 1.22+
- Node 20+
- `yt-dlp` and `ffmpeg` installed and in PATH

### 2) Configure env
Copy `.env.example` to `.env` and adjust as needed.

### 3) Run backend
```bash
cd backend
go mod tidy
go run ./main.go
