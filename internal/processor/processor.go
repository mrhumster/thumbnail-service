//go:generate mockgen -source=processor.go -destination=mock/processor_mock.go -package=mock
package processor

import "context"

type ThumbnailProcessor interface {
	GetDuration(ctx context.Context, inputPath string) (float64, error)
	GenerateThumbnail(ctx context.Context, inputPath, outputPath string, seek float64) error
}
