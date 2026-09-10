package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	pb "github.com/mrhumster/thumbnail-service/gen/go/stream"
	"github.com/mrhumster/thumbnail-service/internal/metrics"
	"github.com/mrhumster/thumbnail-service/internal/processor"
	"github.com/mrhumster/thumbnail-service/internal/storage"
)

const presignTTL = 15 * time.Minute

type HandleThumbnail struct {
	processor     processor.ThumbnailProcessor
	storage       storage.FileStorage
	streamService pb.StreamServiceClient
}

func NewHandleThumbnail(p processor.ThumbnailProcessor, s storage.FileStorage, svc pb.StreamServiceClient) *HandleThumbnail {
	return &HandleThumbnail{
		processor:     p,
		storage:       s,
		streamService: svc,
	}
}

func (h *HandleThumbnail) HandleThumbsnailTask(ctx context.Context, t *asynq.Task) error {
	var p ThumbsnailProcessorPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json unmarshal failed: %v", err)
	}

	start := time.Now()
	taskErr := h.handleThumbnail(ctx, p)
	duration := time.Since(start)

	if taskErr != nil {
		metrics.Errors.Inc()
	} else {
		metrics.Generated.Inc()
		metrics.Duration.Observe(duration.Seconds())
	}
	return taskErr
}

func (h *HandleThumbnail) handleThumbnail(ctx context.Context, p ThumbsnailProcessorPayload) error {

	workDir := fmt.Sprintf("/tmp/%s", p.StreamUUID)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("error creating temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	thumbLocal := filepath.Join(workDir, "thumbnail.jpg")

	slog.Info("generating presigned url", "uuid", p.StreamUUID, "path", p.InputPath)
	inputURL, err := h.storage.GeneratePresignedURL(ctx, p.InputPath, presignTTL)
	if err != nil {
		slog.Error("failed to generate presigned url", "uuid", p.StreamUUID, "error", err)
		return err
	}

	duration, err := h.processor.GetDuration(ctx, inputURL)
	if err != nil {
		slog.Warn("failed to get duration, using 0", "uuid", p.StreamUUID, "error", err)
		duration = 0
	}

	seek := duration * 0.1
	slog.Info("generating thumbnail", "uuid", p.StreamUUID, "duration", duration, "seek", seek)
	if err := h.processor.GenerateThumbnail(ctx, inputURL, thumbLocal, seek); err != nil {
		slog.Error("thumbnail generation failed", "uuid", p.StreamUUID, "error", err)
		if isMissingSource(err) {
			return fmt.Errorf("source missing: %w", asynq.SkipRetry)
		}
		return err
	}

	remoteKey := fmt.Sprintf("thumbnails/%s.jpg", p.StreamUUID)
	slog.Info("uploading thumbnail", "uuid", p.StreamUUID, "key", remoteKey)
	if err := h.storage.Upload(ctx, remoteKey, thumbLocal, "image/jpeg"); err != nil {
		slog.Error("thumbnail upload failed", "uuid", p.StreamUUID, "error", err)
		return err
	}

	if _, err := h.streamService.UpdateStreamProcessing(ctx, &pb.UpdateStreamProcessingRequest{
		StreamUuid: p.StreamUUID.String(),
		Progress:   100,
		Steps:      []string{"Generating thumbnail"},
	}); err != nil {
		slog.Error("grpc update processing failed", "uuid", p.StreamUUID, "error", err)
		return err
	}

	slog.Info("thumbnail generated", "uuid", p.StreamUUID, "key", remoteKey)
	return nil
}

func isMissingSource(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "404") ||
		strings.Contains(msg, "nosuchkey") ||
		strings.Contains(msg, "not found")
}
