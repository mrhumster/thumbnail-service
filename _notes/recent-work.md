# Recent Work

Текущая и недавняя работа над thumbnail-service. Обновляй при значимых изменениях.

## 2026-08-05 — KEDA scale-to-zero для thumbnail-service

- `deploy/k8s/keda/scaledobject.yaml`: `minReplicaCount 1→0` (под существует только при задачах в очереди), `cooldownPeriod 10→120` (под живёт 2 мин после опустошения очереди и подхватывает следующую задачу без холодного старта).
- KEDA scale-to-zero работает через прямое управление replicas Deployment (HPA остаётся `minReplicas: 1`, но KEDA выставляет Deployment replicas в 0).
- **Проверено E2E в кластере:** пустая очередь → 0 подов; тестовая задача `video:thumbsnail` → под 0→1 за ~16с (pollingInterval 30s), превью сгенерировано и залито в MinIO; очередь опустела → через cooldown 120s скалирование в 0. Тестовые объекты в MinIO удалены.

## 2026-08-05 — Превью через presigned URL (без полного скачивания)

- `FileStorage` + `MinIOClient` получили `GeneratePresignedURL` / `PresignedGetObject`; handler работает с URL: ffprobe длительности и ffmpeg seek идут по HTTP Range, видео целиком не скачивается.
- Добавлено логирование ошибок (`slog.Error`) на presign/generate/upload — раньше текст ошибки терялся.
- Missing source детектится по `404`/`NoSuchKey` → `asynq.SkipRetry`.
- Подтверждено локально на Range-сервере: ffprobe+ffmpeg по URL дают валидный JPEG.
- **Проверено в кластере:** тестовый стрим (540 МБ MP4) → превью за ~0.45с (seek 664с по Range), новый JPEG 74КБ залит в MinIO.
- **Инфраструктура:** requests снижены 800m→500m / 512Mi→256Mi (одноузловой kind перегружен по CPU — RollingUpdate surge не помещался).

## 2026-08-05 — Манифесты для деплоя

- Добавлены `Dockerfile` (golang:1.25-alpine builder → alpine+ffmpeg runtime, бинарник `thumbnail-worker`), docker-таргеты в `Makefile` (`docker-build/push/deploy`, `keda-deploy`, `apply-keda`, `logs`).
- `deploy/k8s/thumbnail/` — Deployment `thumbnail-service` (namespace `go-app`, envFrom: thumbnail-service-config + go-app-config + casbin-redis + minio-credentials, emptyDir `/tmp` 10Gi) и ConfigMap (WORKER_CONCURRENCY, WORKER_SHUTDOWN_TIMEOUT). `STREAM_SERVICE_ADDR` берётся из общего `go-app-config`.
- `deploy/k8s/keda/` — ScaledObject `thumbnail-keda` по очереди `thumbsnails` (pending list + active zset, DB 2) + отдельный TriggerAuthentication `thumbnail-keda-redis-auth` (не конфликтует с `keda-redis-auth` от transcoder).

## 2026-08-05 — Реализация сервиса генерации превью (TDD, ветка opencode-test)

- Полная реализация внутренних пакетов по методу TDD: `processor` (ffmpeg/ffprobe), `storage` (MinIO), `queue` (handler `video:thumbsnail`), `worker` (asynq, очередь `"thumbsnails"`).
- Proto скопирован и сгенерирован локально (`make proto` → `gen/go/stream`), `main.go` переведён на локальные импорты, баннер `thumbnail v0.1.0`.
- Контракт с stream-service: `thumbnails/{uuid}.jpg`, JPEG на 10% длительности, gRPC только `UpdateStreamProcessing`.
- Исправлены опечатки конфига (`Passwrod`→`Password`, `StreamSeviceAddr`→`StreamServiceAddr`).
- `go build`, `go vet`, `go test ./...` — всё зелёное.
