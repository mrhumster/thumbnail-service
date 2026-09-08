package main

import (
	"log/slog"
	"os"

	"github.com/common-nighthawk/go-figure"
	"github.com/hibiken/asynq"
	"github.com/mrhumster/thumbnail-service/config"
	pb "github.com/mrhumster/thumbnail-service/gen/go/stream"
	"github.com/mrhumster/thumbnail-service/internal/processor"
	"github.com/mrhumster/thumbnail-service/internal/queue"
	"github.com/mrhumster/thumbnail-service/internal/storage"
	"github.com/mrhumster/thumbnail-service/internal/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	wellcome := figure.NewFigure("thumbnail v0.1.0", "graffiti", true)
	wellcome.Print()
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("error load config")
		os.Exit(1)
	}

	minioStorage, err := storage.NewMinIOStorageFromConfig(cfg.MinIO)
	if err != nil {
		slog.Error("error init minio storage", "error", err)
		os.Exit(1)
	}

	ffmpeg, err := processor.NewFFmpegProcessor()
	if err != nil {
		slog.Error("error init processor", "error", err)
		os.Exit(1)
	}

	conn, err := grpc.NewClient(
		cfg.Server.StreamServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("error init gRPC client: %w", "error", err)
		os.Exit(1)
	}

	defer conn.Close()

	streamServiceClient := pb.NewStreamServiceClient(conn)

	srv := worker.NewAsynqWorker(cfg, streamServiceClient)
	handler := queue.NewHandleThumbnail(ffmpeg, minioStorage, streamServiceClient)
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TaskThumbsnailProcessor, handler.HandleThumbsnailTask)

	slog.Info("Thumbnail Worker started...")
	if err := srv.Run(mux); err != nil {
		slog.Error("could not run asynq server:", "error", err.Error())
		os.Exit(1)
	}
}
