# Memory Index

> Вся память проекта хранится здесь, в `_notes/`. Этот файл — точка входа.
> Содержимое этого файла автоинжектится в контекст каждой новой сессии (см. `opencode.json` → `instructions`).
> При старте сессии: изучи этот хаб, при необходимости открой заметки по ссылкам ниже.

## О проекте

Микросервис генерации превью/тромбнейлов для видеохостинга **GoCast** (`services/thumbnail-service` в монорепо GoCast).
Слушает очередь Asynq, берёт задачи на генерацию превью, обрабатывает видео и складывает результат в MinIO.

## Стек

- Go 1.25.1, модуль `github.com/mrhumster/thumbnail-service`
- Asynq (`github.com/hibiken/asynq`) + Redis — очередь задач
- gRPC-клиент к stream-service (прото `github.com/mrhumster/transcoder-service/gen/go/stream`)
- MinIO — хранение превью

## Команды

```bash
go build ./...        # сборка
go vet ./...          # статическая проверка
```

## Структура (текущая)

- `cmd/worker/main.go` — точка входа Asynq-worker: логирование (slog), конфиг, MinIO storage, ffmpeg, gRPC-клиент к stream-service, регистрация обработчика `video:thumbsnail`.
- `config/config.go` — конфиг из env (REDIS_*, MINIO_*, STREAM_SERVICE_ADDR, WORKER_CONCURRENCY, WORKER_SHUTDOWN_TIMEOUT).
- `internal/processor` — `ThumbnailProcessor` интерфейс + `FFmpegProcessor` (`GetDuration`, `GenerateThumbnail`).
- `internal/storage` — `FileStorage`/`MinIOClient` интерфейсы + `MinIOStorage` + `NewMinIOStorageFromConfig`.
- `internal/queue` — `ThumbsnailProcessorPayload`, `TaskThumbsnailProcessor="video:thumbsnail"`, `HandleThumbnail`.
- `internal/worker` — `RedisOpt`, `asynqConfig`, `NewAsynqWorker` (подписка только на очередь `"thumbsnails"`).
- `gen/go/stream` — сгенерированный gRPC-клиент (make proto).
- `README.md` — спецификация: Asynq, gRPC, Minio.

## Важные наблюдения

- **Контракт с продюсером (stream-service):** тип задачи `video:thumbsnail`, очередь `"thumbsnails"`, payload `{"stream_uuid","input_path"}`, TaskID `thumbs-{uuid}`, MaxRetry(1). Задача уже ставится из `stream-service/internal/queue/task_distributor_impl.go`.
- **Превью:** MinIO ключ `thumbnails/{uuid}.jpg`, JPEG, кадр на 10% длительности, 1280x720. Через gRPC обновляется только `UpdateStreamProcessing` (progress/steps/error), статус READY ставит transcoder — превью его НЕ трогает.
- **Стриминг:** видео целиком НЕ скачивается — ffprobe/ffmpeg работают по presigned GET URL (`GeneratePresignedURL`, 15 мин) с HTTP Range-запросами. `Download` в интерфейсе остался, но handler его не использует.
- **Тесты:** go test, testify + go.uber.org/mock (mockgen), интеграционный тест ffmpeg (skip без бинаря).
- gRPC-протофайлы генерируются локально в `gen/go/stream` под модуль thumbnail-service.

## Заметки

- [Decisions](decisions.md) — лог архитектурных и технических решений.
- [Recent Work](recent-work.md) — текущая и недавняя работа.
- [README.md](../README.md) — спецификация проекта.

## Правила ведения памяти

- Новые факты: обнови `INDEX.md` (добавь тему) и создай/дополни заметку.
- Решения фиксируй в `decisions.md` (контекст → решение → почему).
- Команда `/remember <факт>` сохраняет произвольный факт в память.
- Ссылки между заметками — через `[[wikilinks]]` или относительные пути.
- `_notes/` и `.obsidian/` не коммитятся в git — это локальная память.
