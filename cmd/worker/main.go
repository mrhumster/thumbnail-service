package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/common-nighthawk/go-figure"
	"github.com/hibiken/asynq"
	sharedconfig "github.com/mrhumster/go-shared/config"
	sharedgrpctls "github.com/mrhumster/go-shared/grpctls"
	sharedmetrics "github.com/mrhumster/go-shared/metrics"
	sharedworker "github.com/mrhumster/go-shared/worker"
	pb "github.com/mrhumster/thumbnail-service/gen/go/stream"
	"github.com/mrhumster/thumbnail-service/internal/processor"
	"github.com/mrhumster/thumbnail-service/internal/queue"
	"github.com/mrhumster/thumbnail-service/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	wellcome := figure.NewFigure("thumbnail "+version, "graffiti", true)
	wellcome.Print()
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	slog.SetDefault(logger)

	cfg, err := sharedconfig.LoadConfig()
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

	creds := insecure.NewCredentials()
	if cfg.Server.GRPCTLSEnabled {
		creds, err = sharedgrpctls.ClientTLSCreds(cfg.Server.GRPCTLSCertFile, cfg.Server.GRPCTLSKeyFile, cfg.Server.GRPCTLSCAFile, "stream-service")
		if err != nil {
			slog.Error("error init gRPC TLS client", "error", err)
			os.Exit(1)
		}
	}

	conn, err := grpc.NewClient(
		cfg.Server.StreamServiceAddr,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		slog.Error("error init gRPC client: %w", "error", err)
		os.Exit(1)
	}

	defer conn.Close()

	streamServiceClient := pb.NewStreamServiceClient(conn)

	reportThumbnailError := func(ctx context.Context, task *asynq.Task, err error) {
		var p queue.ThumbsnailProcessorPayload
		if uerr := json.Unmarshal(task.Payload(), &p); uerr != nil {
			return
		}
		streamServiceClient.UpdateStreamProcessing(ctx, &pb.UpdateStreamProcessingRequest{
			StreamUuid: p.StreamUUID.String(),
			Progress:   0,
			Steps:      []string{"Generating thumbnail"},
			Error:      fmt.Sprintf("thumbnail worker failed: %v", err),
		})
	}

	srv, err := sharedworker.NewAsynqServer(sharedworker.Options{
		Addr:            cfg.Redis.Addr,
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		Concurrency:     cfg.Worker.Concurrency,
		ShutdownTimeout: cfg.Worker.ShutdownTimeout,
		Queues:          map[string]int{"thumbsnails": 6},
		ErrorReporter:   reportThumbnailError,
		MetricsAddr:     cfg.Server.MetricsAddr,
	})
	if err != nil {
		slog.Error("error init asynq worker", "error", err)
		os.Exit(1)
	}

	handler := queue.NewHandleThumbnail(ffmpeg, minioStorage, streamServiceClient)
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TaskThumbsnailProcessor, sharedmetrics.Instrument(queue.TaskThumbsnailProcessor, handler.HandleThumbsnailTask))

	slog.Info("Thumbnail Worker started...")
	if err := srv.Run(mux); err != nil {
		slog.Error("could not run asynq server:", "error", err.Error())
		os.Exit(1)
	}
}
