// Purpose: yt-dlp + ffmpeg wrappers (metadata + direct audio URL).

package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	YTDLPPath  string
	FFMPEGPath string
}

type Client struct {
	cfg Config
}

func NewClient(cfg Config) *Client {
	if cfg.YTDLPPath == "" {
		cfg.YTDLPPath = "yt-dlp"
	}
	if cfg.FFMPEGPath == "" {
		cfg.FFMPEGPath = "ffmpeg"
	}
	return &Client{cfg: cfg}
}

type Meta struct {
	Title       string
	DurationSec int
	WebpageURL  string
}

func (c *Client) ResolveMeta(ctx context.Context, url string) (Meta, error) {
	metas, err := c.ResolveMetas(ctx, url)
	if err != nil {
		return Meta{}, err
	}
	if len(metas) == 0 {
		return Meta{}, errors.New("no tracks resolved")
	}
	return metas[0], nil
}

func (c *Client) ResolveMetas(ctx context.Context, url string) ([]Meta, error) {
	// yt-dlp --dump-single-json URL (supports playlists)
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

<<<<<<< ours
	cmd := exec.CommandContext(ctx, c.cfg.YTDLPPath, "--dump-single-json", url)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		return nil, errors.New(strings.TrimSpace(errb.String()))
	}

	var raw map[string]any
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
=======
	out, err := runYTDLP(ctx, c.cfg.YTDLPPath, "--dump-single-json", url)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
>>>>>>> theirs
		return nil, err
	}

	if entries, ok := raw["entries"].([]any); ok {
		metas := make([]Meta, 0, len(entries))
		for _, entry := range entries {
			entryMap, ok := entry.(map[string]any)
			if !ok || len(entryMap) == 0 {
				continue
			}
			meta := metaFromMap(entryMap, url)
			metas = append(metas, meta)
		}
		return metas, nil
	}

	return []Meta{metaFromMap(raw, url)}, nil
}

func metaFromMap(raw map[string]any, fallbackURL string) Meta {
	title, _ := raw["title"].(string)
	webpage, _ := raw["webpage_url"].(string)
	id, _ := raw["id"].(string)
	urlValue, _ := raw["url"].(string)

	dur := 0
	switch v := raw["duration"].(type) {
	case float64:
		dur = int(v)
	case string:
		if n, _ := strconv.Atoi(v); n > 0 {
			dur = n
		}
	}

	if title == "" {
		title = "Unknown title"
	}
	if webpage == "" {
		webpage = normalizeURL(urlValue, id, fallbackURL)
	}

	return Meta{Title: title, DurationSec: dur, WebpageURL: webpage}
}

func normalizeURL(urlValue, id, fallbackURL string) string {
	if strings.HasPrefix(urlValue, "http://") || strings.HasPrefix(urlValue, "https://") {
		return urlValue
	}
	if id != "" {
		return "https://www.youtube.com/watch?v=" + id
	}
	if urlValue != "" {
		return "https://www.youtube.com/watch?v=" + urlValue
	}
	return fallbackURL
}

func (c *Client) DirectAudioURL(ctx context.Context, url string) (string, error) {
	// yt-dlp -f bestaudio -g URL  => prints direct media URL
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	out, err := runYTDLP(ctx, c.cfg.YTDLPPath, "-f", "bestaudio", "-g", "--no-playlist", url)
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return "", errors.New("yt-dlp returned empty direct url")
	}
	return strings.TrimSpace(lines[0]), nil
}

func (c *Client) FFMPEGPath() string {
	return c.cfg.FFMPEGPath
}

func runYTDLP(ctx context.Context, path string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, path, args...)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, errors.New(msg)
	}
	return out.Bytes(), nil
}
