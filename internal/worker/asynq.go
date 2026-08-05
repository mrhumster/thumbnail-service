package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/mrhumster/thumbnail-service/config"
	pb "github.com/mrhumster/thumbnail-service/gen/go/stream"
	"github.com/mrhumster/thumbnail-service/internal/queue"
)

func RedisOpt(c config.Redis) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       c.DB,
	}
}

func asynqConfig(c *config.Config, streamSvc pb.StreamServiceClient) asynq.Config {
	return asynq.Config{
		Concurrency:     c.Worker.Concurrency,
		ShutdownTimeout: c.Worker.ShutdownTimeout,
		Queues:          map[string]int{"thumbsnails": 6},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			var p queue.ThumbsnailProcessorPayload
			if err := json.Unmarshal(task.Payload(), &p); err != nil {
				return
			}
			streamSvc.UpdateStreamProcessing(ctx, &pb.UpdateStreamProcessingRequest{
				StreamUuid: p.StreamUUID.String(),
				Progress:   0,
				Steps:      []string{"Generating thumbnail"},
				Error:      fmt.Sprintf("thumbnail worker failed: %v", err),
			})
		}),
	}
}

func NewAsynqWorker(c *config.Config, streamSvc pb.StreamServiceClient) *asynq.Server {
	return asynq.NewServer(RedisOpt(c.Redis), asynqConfig(c, streamSvc))
}
