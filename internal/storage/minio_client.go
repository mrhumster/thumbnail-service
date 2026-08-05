//go:generate mockgen -source=minio_client.go -destination=mock/minio_client_mock.go -package=mock
package storage

import (
	"context"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinIOClient interface {
	FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts minio.GetObjectOptions) error
	FPutObject(ctx context.Context, bucketName, objectName, filePath string, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	PresignedGetObject(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values) (*url.URL, error)
}
