package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/mrhumster/thumbnail-service/gen/go/stream"
	mockProc "github.com/mrhumster/thumbnail-service/internal/processor/mock"
	mockSvc "github.com/mrhumster/thumbnail-service/internal/service/mock"
	mockStor "github.com/mrhumster/thumbnail-service/internal/storage/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestTask(t *testing.T, payload ThumbsnailProcessorPayload) *asynq.Task {
	t.Helper()
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)
	return asynq.NewTask(TaskThumbsnailProcessor, payloadBytes)
}

func newHandler(ctrl *gomock.Controller) (*HandleThumbnail, *mockProc.MockThumbnailProcessor, *mockStor.MockFileStorage, *mockSvc.MockStreamServiceClient) {
	mockProcessor := mockProc.NewMockThumbnailProcessor(ctrl)
	mockStorage := mockStor.NewMockFileStorage(ctrl)
	mockService := mockSvc.NewMockStreamServiceClient(ctrl)
	handler := NewHandleThumbnail(mockProcessor, mockStorage, mockService)
	return handler, mockProcessor, mockStorage, mockService
}

func TestHandleThumbnail_HandleThumbsnailTask(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		handler, mockProcessor, mockStorage, mockService := newHandler(ctrl)
		ctx := context.Background()
		streamUUID := uuid.New()
		task := newTestTask(t, ThumbsnailProcessorPayload{
			StreamUUID: streamUUID,
			InputPath:  "raw/video.mp4",
		})

		mockStorage.EXPECT().
			GeneratePresignedURL(gomock.Any(), "raw/video.mp4", presignTTL).
			Return("http://minio.local/raw/video.mp4?X-Amz-Expires=900", nil)
		mockProcessor.EXPECT().
			GetDuration(gomock.Any(), "http://minio.local/raw/video.mp4?X-Amz-Expires=900").
			Return(30.0, nil)
		mockProcessor.EXPECT().
			GenerateThumbnail(gomock.Any(), "http://minio.local/raw/video.mp4?X-Amz-Expires=900", gomock.Any(), 3.0).
			Return(nil)
		mockStorage.EXPECT().
			Upload(gomock.Any(), fmt.Sprintf("thumbnails/%s.jpg", streamUUID), gomock.Any(), "image/jpeg").
			Return(nil)
		mockService.EXPECT().
			UpdateStreamProcessing(gomock.Any(), gomock.Any()).
			Return(&stream.UpdateStreamProcessingResponse{Updated: true}, nil)

		err := handler.HandleThumbsnailTask(ctx, task)
		require.NoError(t, err)
	})

	t.Run("invalid payload", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		handler, _, _, _ := newHandler(ctrl)
		ctx := context.Background()
		task := asynq.NewTask(TaskThumbsnailProcessor, []byte("not-json"))

		err := handler.HandleThumbsnailTask(ctx, task)
		require.Error(t, err)
	})

	t.Run("presign error propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		handler, _, mockStorage, _ := newHandler(ctrl)
		ctx := context.Background()
		task := newTestTask(t, ThumbsnailProcessorPayload{
			StreamUUID: uuid.New(),
			InputPath:  "raw/video.mp4",
		})

		mockStorage.EXPECT().
			GeneratePresignedURL(gomock.Any(), "raw/video.mp4", presignTTL).
			Return("", fmt.Errorf("presign error"))

		err := handler.HandleThumbsnailTask(ctx, task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "presign error")
	})

	t.Run("missing source skips retry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		handler, mockProcessor, mockStorage, mockService := newHandler(ctrl)
		ctx := context.Background()
		task := newTestTask(t, ThumbsnailProcessorPayload{
			StreamUUID: uuid.New(),
			InputPath:  "raw/missing.mp4",
		})

		mockStorage.EXPECT().
			GeneratePresignedURL(gomock.Any(), "raw/missing.mp4", presignTTL).
			Return("http://minio.local/raw/missing.mp4?X-Amz-Expires=900", nil)
		mockProcessor.EXPECT().
			GetDuration(gomock.Any(), gomock.Any()).
			Return(0.0, fmt.Errorf("ffprobe 404"))
		mockProcessor.EXPECT().
			GenerateThumbnail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("ffmpeg: 404 Not Found"))
		mockService.EXPECT().
			UpdateStreamProcessing(gomock.Any(), gomock.Any()).
			Return(&stream.UpdateStreamProcessingResponse{Updated: true}, nil)

		err := handler.HandleThumbsnailTask(ctx, task)
		require.Error(t, err)
		assert.ErrorIs(t, err, asynq.SkipRetry)
	})

	t.Run("generate error propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		handler, mockProcessor, mockStorage, mockService := newHandler(ctrl)
		ctx := context.Background()
		task := newTestTask(t, ThumbsnailProcessorPayload{
			StreamUUID: uuid.New(),
			InputPath:  "raw/video.mp4",
		})

		mockStorage.EXPECT().
			GeneratePresignedURL(gomock.Any(), "raw/video.mp4", presignTTL).
			Return("http://minio.local/raw/video.mp4", nil)
		mockProcessor.EXPECT().
			GetDuration(gomock.Any(), gomock.Any()).
			Return(30.0, nil)
		mockProcessor.EXPECT().
			GenerateThumbnail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("generate error"))
		mockService.EXPECT().
			UpdateStreamProcessing(gomock.Any(), gomock.Any()).
			Return(&stream.UpdateStreamProcessingResponse{Updated: true}, nil)

		err := handler.HandleThumbsnailTask(ctx, task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "generate error")
	})

	t.Run("upload error propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		handler, mockProcessor, mockStorage, _ := newHandler(ctrl)
		ctx := context.Background()
		task := newTestTask(t, ThumbsnailProcessorPayload{
			StreamUUID: uuid.New(),
			InputPath:  "raw/video.mp4",
		})

		mockStorage.EXPECT().
			GeneratePresignedURL(gomock.Any(), "raw/video.mp4", presignTTL).
			Return("http://minio.local/raw/video.mp4", nil)
		mockProcessor.EXPECT().
			GetDuration(gomock.Any(), gomock.Any()).
			Return(30.0, nil)
		mockProcessor.EXPECT().
			GenerateThumbnail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		mockStorage.EXPECT().
			Upload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("upload error"))

		err := handler.HandleThumbsnailTask(ctx, task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "upload error")
	})

	t.Run("grpc update error propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		handler, mockProcessor, mockStorage, mockService := newHandler(ctrl)
		ctx := context.Background()
		task := newTestTask(t, ThumbsnailProcessorPayload{
			StreamUUID: uuid.New(),
			InputPath:  "raw/video.mp4",
		})

		mockStorage.EXPECT().
			GeneratePresignedURL(gomock.Any(), "raw/video.mp4", presignTTL).
			Return("http://minio.local/raw/video.mp4", nil)
		mockProcessor.EXPECT().
			GetDuration(gomock.Any(), gomock.Any()).
			Return(30.0, nil)
		mockProcessor.EXPECT().
			GenerateThumbnail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		mockStorage.EXPECT().
			Upload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		mockService.EXPECT().
			UpdateStreamProcessing(gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("grpc error"))

		err := handler.HandleThumbsnailTask(ctx, task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "grpc error")
	})
}
