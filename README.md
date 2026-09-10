# thumbnail-service

Generates video preview thumbnails for GoCast streams. An **asynq** worker scaled by
**KEDA** — it sleeps at 0 replicas and wakes up when a `thumbsnails` task lands in the queue.

## How it works

- Stream-service enqueues a `thumbsnails` task (Redis asynq queue, DB 2);
- KEDA `ScaledObject` watches the queue length and scales the deployment from 0;
- The worker reads the source video from MinIO, extracts a frame with ffmpeg, writes the
  thumbnail, and reports progress/state via gRPC `UpdateStreamProcessing` back to stream-service;
- KEDA scales back to 0 when the queue drains (normal).

## Spec

| Layer | Tech |
|---|---|
| Task queue | Asynq (Redis), queue `thumbsnails` |
| Processing | ffmpeg (`internal/processor`) |
| Storage | MinIO (`internal/storage`) |
| gRPC | Client to stream-service (**mTLS**, serverName `stream-service`) |
| Metrics | Prometheus `/metrics` on `METRICS_ADDR` (default `:9090` in K8s) |
| Config | `sharedconfig` from `go-shared` (env-driven) |

## Metrics

Exposed via `METRICS_ADDR` (empty = off) with `promhttp`:

- `thumbnail_generated_total`, `thumbnail_generation_duration_seconds`,
  `thumbnail_generation_errors_total`

Plus shared asynq-task metrics from `go-shared/metrics` (`asynq_task_processed_total`,
`asynq_task_duration_seconds`, `asynq_task_inflight`). Kubernetes liveness probes hit
`/metrics` (HTTP GET) instead of `ps`, and the Deployment carries `prometheus.io/*` annotations.

## Configuration

Loaded from env by `sharedconfig.LoadConfig()` (root `.env` → `thumbnail-service-config`
ConfigMap). See `services/shared/README.md` for the full variable list; the relevant ones:
`REDIS_ADDR`, `MINIO_ENDPOINT`/`MINIO_ACCESS_KEY`/`MINIO_SECRET_KEY`/`MINIO_BUCKET_NAME`,
`STREAM_SERVICE_ADDRESS`, `GRPC_TLS_*`, `METRICS_ADDR`.

## Deployment

K8s manifests under `deploy/k8s/`:

```
deploy/k8s/
├── keda/
│   ├── auth.yaml           # Redis trigger auth
│   └── scaledobject.yaml   # scale from/to 0 based on queue length
└── thumbnail/
    ├── deployment.yaml     # image xomrkob/thumbnail-service:<git-tag>, metrics containerPort
    └── service.yaml        # ClusterIP for :9090 metrics scraping
```

Build/push/deploy: `make build push deploy` (image `xomrkob/thumbnail-service:<git-tag>`).

The asynq worker scaffolding (`worker.NewAsynqServer`, ErrorHandler wiring) is shared with
transcoder — see `services/shared/worker`.