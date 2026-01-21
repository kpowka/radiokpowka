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
	// yt-dlp --dump-json URL
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.cfg.YTDLPPath, "--dump-json", "--no-playlist", url)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		return Meta{}, errors.New(strings.TrimSpace(errb.String()))
	}

	var raw map[string]any
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		return Meta{}, err
	}

	title, _ := raw["title"].(string)
	webpage, _ := raw["webpage_url"].(string)

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
		webpage = url
	}

	return Meta{Title: title, DurationSec: dur, WebpageURL: webpage}, nil
}

func (c *Client) DirectAudioURL(ctx context.Context, url string) (string, error) {
	// yt-dlp -f bestaudio -g URL  => prints direct media URL
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.cfg.YTDLPPath, "-f", "bestaudio", "-g", "--no-playlist", url)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		return "", errors.New(strings.TrimSpace(errb.String()))
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return "", errors.New("yt-dlp returned empty direct url")
	}
	return strings.TrimSpace(lines[0]), nil
}

func (c *Client) FFMPEGPath() string {
	return c.cfg.FFMPEGPath
}
