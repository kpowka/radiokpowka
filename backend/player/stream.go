// Purpose: Streaming implementation for /stream.
// - If paused: stream silence from ffmpeg (anullsrc) to avoid client error loops.
// - If playing: get yt-dlp direct audio URL and run ffmpeg -> mp3, seek to position.

package player

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
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
		"-loglevel", "warning",
		"-fflags", "+flush_packets",
		"-f", "lavfi",
		"-i", "anullsrc=r=48000:cl=stereo",
		"-acodec", "libmp3lame",
		"-b:a", "192k",
		"-flush_packets", "1",
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
		"-loglevel", "warning",
		"-fflags", "+flush_packets",
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
		"-flush_packets", "1",
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
	fw := &flushWriter{w: w, f: flusher}

	// no track -> silence for a short time then exit
	if !ok || url == "" {
		shortCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		log.Printf("стрим: нет трека, отправляем короткую тишину")
		counter := &countWriter{w: fw}
		cmd := NewSilenceFFMPEG(shortCtx, ffmpegPath, counter)
		return runFFMPEG(shortCtx, cmd, "silence", counter)
	}

	// paused => stream silence (keeps clients stable)
	if paused {
		log.Printf("стрим: пауза, отправляем тишину")
		counter := &countWriter{w: fw}
		cmd := NewSilenceFFMPEG(ctx, ffmpegPath, counter)
		return runFFMPEG(ctx, cmd, "silence", counter)
	}

	// playing => direct URL via yt-dlp, then ffmpeg to mp3
	direct, err := c.yt.DirectAudioURL(ctx, url)
	if err != nil {
		log.Printf("стрим: ошибка получения direct URL: %v", err)
		return err
	}

	log.Printf("стрим: старт трека url=%s pos=%d vol=%.3f", url, pos, vol)
	counter := &countWriter{w: fw}
	cmd := NewTrackFFMPEG(ctx, ffmpegPath, direct, pos, vol, counter)
	return runFFMPEG(ctx, cmd, "track", counter)
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

type countWriter struct {
	w io.Writer
	n int64
}

func (cw *countWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	atomic.AddInt64(&cw.n, int64(n))
	return n, err
}

func (cw *countWriter) Count() int64 {
	return atomic.LoadInt64(&cw.n)
}

func runFFMPEG(ctx context.Context, cmd *exec.Cmd, label string, counter *countWriter) error {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("ffmpeg %s stderr: %s", label, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Printf("ffmpeg %s stderr ошибка чтения: %v", label, err)
		}
	}()

	err = cmd.Wait()
	wg.Wait()

	bytesSent := int64(0)
	if counter != nil {
		bytesSent = counter.Count()
	}
	if bytesSent == 0 {
		log.Printf("ffmpeg %s: предупреждение, отправлено 0 байт", label)
	} else {
		log.Printf("ffmpeg %s: отправлено байт=%d", label, bytesSent)
	}

	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
		log.Printf("ffmpeg %s завершился с ошибкой: %v", label, err)
		return err
	}
	return nil
}

// ensure we use uuid import (kept for future extensions, avoids IDE confusion)
var _ = uuid.Nil
var _ = youtube.Meta{}
