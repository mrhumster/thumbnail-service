package storage

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

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

func TestMinioStorage_Upload(t *testing.T) {	t.Run("success", func(t *testing.T) {
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

func TestMinioStorage_GeneratePresignedURL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockMinioClient := mock.NewMockMinIOClient(ctrl)
		s := NewMinIOStorage(mockMinioClient, "files")
		ctx := context.Background()

		u, err := url.Parse("https://minio:9000/files/raw/video.mp4?X-Amz-Expires=900")
		require.NoError(t, err)
		mockMinioClient.EXPECT().
			PresignedGetObject(gomock.Any(), "files", "raw/video.mp4", 15*time.Minute, gomock.Any()).
			Return(u, nil)

		got, err := s.GeneratePresignedURL(ctx, "raw/video.mp4", 15*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, u.String(), got)
	})

	t.Run("client error propagation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockMinioClient := mock.NewMockMinIOClient(ctrl)
		s := NewMinIOStorage(mockMinioClient, "files")
		ctx := context.Background()

		mockMinioClient.EXPECT().
			PresignedGetObject(gomock.Any(), "files", "raw/video.mp4", 15*time.Minute, gomock.Any()).
			Return(nil, fmt.Errorf("presign error"))

		_, err := s.GeneratePresignedURL(ctx, "raw/video.mp4", 15*time.Minute)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "presign error")
	})
}
