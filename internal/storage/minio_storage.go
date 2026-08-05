package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinIOStorage struct {
	Client MinIOClient
	Bucket string
}

func NewMinIOStorage(client MinIOClient, bucket string) *MinIOStorage {
	return &MinIOStorage{
		Client: client,
		Bucket: bucket,
	}
}

func (s *MinIOStorage) Download(ctx context.Context, objectName, filePath string) error {
	if err := s.Client.FGetObject(ctx, s.Bucket, objectName, filePath, minio.GetObjectOptions{}); err != nil {
		return fmt.Errorf("error download from storage: %w", err)
	}
	return nil
}

func (s *MinIOStorage) Upload(ctx context.Context, objectName, filePath, contentType string) error {
	uploadInfo, err := s.Client.FPutObject(ctx, s.Bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("error uploading to bucket %s %v", s.Bucket, err)
	}
	slog.Info("Upload success", "Bucket", s.Bucket, "uploadInfo", uploadInfo)
	return nil
}

func (s *MinIOStorage) GeneratePresignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	u, err := s.Client.PresignedGetObject(ctx, s.Bucket, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("error generating presigned url for %s: %w", objectName, err)
	}
	return u.String(), nil
}
