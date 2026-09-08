package processor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestVideo(t *testing.T, dir string) string {
	t.Helper()
	ffmpegBin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not found, skipping integration test")
	}
	path := filepath.Join(dir, "fixture.mp4")
	cmd := exec.Command(ffmpegBin, "-y", "-f", "lavfi",
		"-i", "testsrc=duration=2:size=320x240:rate=10",
		"-f", "mp4", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("cannot generate fixture video: %v: %s", err, out)
	}
	return path
}

func TestFFmpegProcessor(t *testing.T) {
	ctx := context.Background()
	proc, err := NewFFmpegProcessor()
	require.NoError(t, err)

	dir := t.TempDir()
	videoPath := createTestVideo(t, dir)

	t.Run("GetDuration returns video duration", func(t *testing.T) {
		duration, err := proc.GetDuration(ctx, videoPath)
		require.NoError(t, err)
		assert.InDelta(t, 2.0, duration, 0.5)
	})

	t.Run("GenerateThumbnail produces a valid JPEG", func(t *testing.T) {
		outPath := filepath.Join(dir, "thumb.jpg")
		err := proc.GenerateThumbnail(ctx, videoPath, outPath, 0.2)
		require.NoError(t, err)

		data, err := os.ReadFile(outPath)
		require.NoError(t, err)
		assert.True(t, len(data) > 0, "thumbnail should not be empty")
		assert.Equal(t, []byte{0xFF, 0xD8, 0xFF}, data[:3], "thumbnail must be a JPEG")
	})

	t.Run("GenerateThumbnail errors on missing input", func(t *testing.T) {
		err := proc.GenerateThumbnail(ctx, filepath.Join(dir, "missing.mp4"), filepath.Join(dir, "nope.jpg"), 1.0)
		require.Error(t, err)
	})

	t.Run("GetDuration errors on missing input", func(t *testing.T) {
		_, err := proc.GetDuration(ctx, filepath.Join(dir, "missing.mp4"))
		require.Error(t, err)
	})
}
