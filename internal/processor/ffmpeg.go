package processor

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type FFmpegProcessor struct {
	binPath string
}

func NewFFmpegProcessor() (*FFmpegProcessor, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found in system: %w", err)
	}
	return &FFmpegProcessor{binPath: path}, nil
}

func (p *FFmpegProcessor) GetDuration(ctx context.Context, inputPath string) (float64, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	}
	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}
	durationStr := strings.TrimSpace(string(out))
	return strconv.ParseFloat(durationStr, 64)
}

func (p *FFmpegProcessor) GenerateThumbnail(ctx context.Context, inputPath, outputPath string, seek float64) error {
	args := []string{
		"-ss", strconv.FormatFloat(seek, 'f', 2, 64),
		"-i", inputPath,
		"-frames:v", "1",
		"-vf", "scale=1280:-2",
		"-q:v", "3",
		"-y",
		outputPath,
	}
	cmd := exec.CommandContext(ctx, p.binPath, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg thumbnail generation failed: %w: %s", err, out)
	}
	return nil
}
