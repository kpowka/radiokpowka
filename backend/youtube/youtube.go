// Purpose: yt-dlp + ffmpeg wrappers (metadata + direct audio URL).

package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	YTDLPPath          string
	FFMPEGPath         string
	CookiesFromBrowser string
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
	if cfg.CookiesFromBrowser == "" {
		log.Printf("ПРЕДУПРЕЖДЕНИЕ: YTDLP_COOKIES_FROM_BROWSER не задан, возможны ошибки авторизации YouTube.")
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
	// Prefer --dump-single-json if supported, fallback to --dump-json lines.
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	if out, err := c.runYTDLP(ctx, "--dump-single-json", url); err == nil {
		if metas, err := parseMetasFromJSON(out, url); err == nil && len(metas) > 0 {
			return metas, nil
		}
	}

	out, err := c.runYTDLP(ctx, "--dump-json", url)
	if err != nil {
		return nil, err
	}

	metas, err := parseMetasFromLines(out, url)
	if err != nil {
		return nil, err
	}
	if len(metas) == 0 {
		return nil, errors.New("no tracks resolved")
	}
	return metas, nil
}

func parseMetasFromJSON(out []byte, fallbackURL string) ([]Meta, error) {
	var raw map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	return parseMetasFromMap(raw, fallbackURL), nil
}

func parseMetasFromLines(out []byte, fallbackURL string) ([]Meta, error) {
	entries := strings.Split(strings.TrimSpace(string(out)), "\n")
	metas := make([]Meta, 0, len(entries))
	for _, line := range entries {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, err
		}
		metas = append(metas, parseMetasFromMap(raw, fallbackURL)...)
	}
	return metas, nil
}

func parseMetasFromMap(raw map[string]any, fallbackURL string) []Meta {
	if nested, ok := raw["entries"].([]any); ok {
		metas := make([]Meta, 0, len(nested))
		for _, entry := range nested {
			entryMap, ok := entry.(map[string]any)
			if !ok || len(entryMap) == 0 {
				continue
			}
			meta := metaFromMap(entryMap, fallbackURL)
			metas = append(metas, meta)
		}
		return metas
	}
	return []Meta{metaFromMap(raw, fallbackURL)}
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

	out, err := c.runYTDLP(ctx, "-f", "bestaudio", "-g", "--no-playlist", url)
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

func (c *Client) runYTDLP(ctx context.Context, args ...string) ([]byte, error) {
	fullArgs := make([]string, 0, len(args)+2)
	if c.cfg.CookiesFromBrowser != "" {
		fullArgs = append(fullArgs, "--cookies-from-browser", c.cfg.CookiesFromBrowser)
	}
	fullArgs = append(fullArgs, args...)

	cmd := exec.CommandContext(ctx, c.cfg.YTDLPPath, fullArgs...)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		stdoutText := strings.TrimSpace(out.String())
		stderrText := strings.TrimSpace(errb.String())
		log.Printf("yt-dlp stdout: %s", trimForLog(stdoutText))
		log.Printf("yt-dlp stderr: %s", trimForLog(stderrText))
		msg := stderrText
		if msg == "" {
			msg = err.Error()
		}
		return nil, errors.New(msg)
	}
	stdoutText := strings.TrimSpace(out.String())
	stderrText := strings.TrimSpace(errb.String())
	log.Printf("yt-dlp stdout: %s", trimForLog(stdoutText))
	log.Printf("yt-dlp stderr: %s", trimForLog(stderrText))
	return out.Bytes(), nil
}

func trimForLog(s string) string {
	const maxLen = 2000
	if s == "" {
		return "<пусто>"
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(обрезано)"
}
