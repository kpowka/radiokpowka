// Purpose: Streaming implementation for /stream.
// - If paused: stream silence from ffmpeg (anullsrc) to avoid client error loops.
// - If playing: get yt-dlp direct audio URL and run ffmpeg -> mp3, seek to position.

package player

import (
	"context"
	"io"
	"os/exec"
	"strconv"
	"time"

	"github.com/google/uuid"

	"radiokpowka/backend/youtube"
)

func NewSilenceFFMPEG(ctx context.Context, ffmpegPath string, w io.Writer) *exec.Cmd {
	// Generates infinite silent MP3 stream
	// ffmpeg -f lavfi -i anullsrc=r=48000:cl=stereo -acodec libmp3lame -b:a 192k -f mp3 pipe:1
	cmd := exec.CommandContext(ctx,
		ffmpegPath,
		"-hide_banner",
		"-loglevel", "error",
		"-f", "lavfi",
		"-i", "anullsrc=r=48000:cl=stereo",
		"-acodec", "libmp3lame",
		"-b:a", "192k",
		"-f", "mp3",
		"pipe:1",
	)
	cmd.Stdout = w
	return cmd
}

func NewTrackFFMPEG(ctx context.Context, ffmpegPath string, directURL string, seekSec int, volume float64, w io.Writer) *exec.Cmd {
	// ffmpeg -ss <seek> -i <directURL> -vn -filter:a volume=<v> -acodec libmp3lame -b:a 192k -f mp3 pipe:1
	args := []string{
		"-hide_banner",
		"-loglevel", "error",
	}
	if seekSec > 0 {
		args = append(args, "-ss", strconv.Itoa(seekSec))
	}
	args = append(args,
		"-i", directURL,
		"-vn",
		"-filter:a", "volume="+formatVol(volume),
		"-acodec", "libmp3lame",
		"-b:a", "192k",
		"-f", "mp3",
		"pipe:1",
	)
	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	cmd.Stdout = w
	return cmd
}

func formatVol(v float64) string {
	if v < 0 {
		v = 0
	}
	if v > 2 {
		v = 2
	}
	// ffmpeg accepts float in string form
	return strconv.FormatFloat(v, 'f', 3, 64)
}

type ControllerStreamer interface {
	CurrentForStreaming() (url string, posSec int, vol float64, paused bool, ok bool)
}

type StreamWriter interface {
	Write([]byte) (int, error)
}

type Flusher interface {
	Flush()
}

func (c *Controller) StreamTo(ctx context.Context, w io.Writer, flusher Flusher) error {
	url, pos, vol, paused, ok := c.CurrentForStreaming()
	ffmpegPath := c.yt.FFMPEGPath()

	// no track -> silence for a short time then exit
	if !ok || url == "" {
		shortCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		cmd := NewSilenceFFMPEG(shortCtx, ffmpegPath, &flushWriter{w: w, f: flusher})
		_ = cmd.Run()
		return nil
	}

	// paused => stream silence (keeps clients stable)
	if paused {
		cmd := NewSilenceFFMPEG(ctx, ffmpegPath, &flushWriter{w: w, f: flusher})
		_ = cmd.Run()
		return nil
	}

	// playing => direct URL via yt-dlp, then ffmpeg to mp3
	direct, err := c.yt.DirectAudioURL(ctx, url)
	if err != nil {
		return err
	}

	cmd := NewTrackFFMPEG(ctx, ffmpegPath, direct, pos, vol, &flushWriter{w: w, f: flusher})
	_ = cmd.Run()
	return nil
}

type flushWriter struct {
	w io.Writer
	f Flusher
}

func (fw *flushWriter) Write(p []byte) (int, error) {
	n, err := fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}
	return n, err
}

// ensure we use uuid import (kept for future extensions, avoids IDE confusion)
var _ = uuid.Nil
var _ = youtube.Meta{}
