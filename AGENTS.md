# AGENTS.md

## Project Memory

Память проекта хранится в Obsidian vault (корень проекта), заметки — в `_notes/`.

- **При старте сессии контекст восстанавливается из `_notes/INDEX.md`** — он автоинжектится через `opencode.json` → `instructions`. По задаче открывай нужные заметки по ссылкам из хаба.
- Новые факты/решения сохраняй в `_notes/`: обнови `INDEX.md` и соответствующую заметку (решения — в `decisions.md`).
- Команда `/remember <факт>` сохраняет произвольный факт в память.
- `_notes/` и `.obsidian/` не коммитятся в git — это локальная память.

## Commands

```bash
go build ./...   # сборка
go vet ./...     # статическая проверка
```

## Architecture

Go worker-сервис генерации превью для GoCast.

- **Очередь:** Asynq + Redis (REDIS_* env), обработчик задач в `cmd/worker/main.go`.
- **Хранилище:** MinIO (MINIO_* env) — сюда складываются превью.
- **Интеграция:** gRPC-клиент к stream-service (`STREAM_SERVICE_ADDR`, прото `transcoder-service/gen/go/stream`).
- **Конфиг:** `config/config.go` — env-based, уже реализован.
- **TODO:** internal-пакеты `processor` (ffmpeg), `storage` (minio), `queue` (handlers), `worker` (asynq) — сейчас `main.go` импортирует их из `github.com/mrhumster/transcoder-service`.
