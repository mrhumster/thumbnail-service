package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/mrhumster/thumbnail-service/internal/storage/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMinioStorage_Download(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockMinioClient := mock.NewMockMinIOClient(ctrl)
		s := NewMinIOStorage(mockMinioClient, "files")
		ctx := context.Background()
		mockMinioClient.EXPECT().
			FGetObject(gomock.Any(), "files", "raw/video.mp4", "/tmp/in.mp4", gomock.Any()).
			Return(nil)
		err := s.Download(ctx, "raw/video.mp4", "/tmp/in.mp4")
		require.NoError(t, err)
	})

	t.Run("client error propagation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockMinioClient := mock.NewMockMinIOClient(ctrl)
		s := NewMinIOStorage(mockMinioClient, "files")
		ctx := context.Background()
		mockMinioClient.EXPECT().
			FGetObject(gomock.Any(), "files", "raw/video.mp4", "/tmp/in.mp4", gomock.Any()).
			Return(fmt.Errorf("client error"))
		err := s.Download(ctx, "raw/video.mp4", "/tmp/in.mp4")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client error")
	})
}

func TestMinioStorage_Upload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockMinioClient := mock.NewMockMinIOClient(ctrl)
		s := NewMinIOStorage(mockMinioClient, "files")
		ctx := context.Background()
		mockMinioClient.EXPECT().
			FPutObject(gomock.Any(), "files", "thumbnails/thumb.jpg", "/tmp/thumb.jpg", gomock.Any()).
			Return(minio.UploadInfo{}, nil)
		err := s.Upload(ctx, "thumbnails/thumb.jpg", "/tmp/thumb.jpg", "image/jpeg")
		require.NoError(t, err)
	})

	t.Run("client error propagation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockMinioClient := mock.NewMockMinIOClient(ctrl)
		s := NewMinIOStorage(mockMinioClient, "files")
		ctx := context.Background()
		mockMinioClient.EXPECT().
			FPutObject(gomock.Any(), "files", "thumbnails/thumb.jpg", "/tmp/thumb.jpg", gomock.Any()).
			Return(minio.UploadInfo{}, fmt.Errorf("minio error"))
		err := s.Upload(ctx, "thumbnails/thumb.jpg", "/tmp/thumb.jpg", "image/jpeg")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "minio error")
	})
}
