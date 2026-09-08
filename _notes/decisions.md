# Decisions

Лог решений по проекту. Формат: **Дата — Решение — Контекст — Почему**.

## 2026-08-05 — Память проекта хранится в Obsidian vault

- **Контекст:** нужна персистентная память, чтобы контекст переживал перезапуски сессий.
- **Решение:** vault Obsidian = корень проекта; заметки памяти в `_notes/`; правила описаны в `AGENTS.md`; `opencode.json` автоинжектит `INDEX.md` при старте сессии.
- **Почему:** обычные `.md` файлы, не требует запущенного Obsidian, не требует MCP/плагинов, `.obsidian/` и `_notes/` не коммитятся в git. Тот же паттерн уже используется в `services/web-frontend`.

## 2026-08-05 — Дизайн thumbnail-service

- **Контекст:** сервис как transcoder, но генерирует превью (картинку) по задаче из очереди. Продюсер задач уже существует: stream-service ставит `video:thumbsnail` в очередь `"thumbsnails"` с payload `{stream_uuid, input_path}`.
- **Решение:**
  - В proto `UpdateStreamMetadataRequest` поля для превью нет; НЕ расширяем proto и stream-service. Превью пишется по фиксированному ключу MinIO `thumbnails/{uuid}.jpg`, URL строится по конвенции. Через gRPC обновляется только `UpdateStreamProcessing` (progress/steps/error); статус `READY` превью не ставит (его ставит transcoder после HLS — иначе превью пометило бы стрим готовым раньше времени).
  - Генерация: один JPEG-кадр на 10% длительности (`ffmpeg -ss <seek> -frames:v 1 -vf scale=1280:-2 -q:v 3`). Бизнес-правило «10%» живёт в handler (`seek := duration*0.1`), processor — чистый экстрактор кадра по заданному seek.
  - Asynq-сервер подписан ТОЛЬКО на очередь `"thumbsnails"` (`Queues: {"thumbsnails": 6}`).
  - Исправлены опечатки конфига: `Passwrod`→`Password`, `StreamSeviceAddr`→`StreamServiceAddr` (env-ключи не менялись).
- **Почему:** минимальная связность с stream-service (не нужен PR туда), конвенция ключей совпадает с raw/ и processed/ паттерном; TDD — каждый пакет покрыт юнит-тестами (testify + go.uber.org/mock), ffmpeg-реализация — интеграционным тестом (skip без ffmpeg).

## 2026-08-05 — Превью через presigned URL вместо полного скачивания

- **Контекст:** thumbnail-service качал исходный MP4 целиком (`FGetObject` на диск), затем отдавал локальный путь ffmpeg. Для большого видео это сотни МБ трафика впустую ради одного кадра.
- **Решение:** ffmpeg/ffprobe работают напрямую с **presigned GET URL**: в `FileStorage` добавлен `GeneratePresignedURL` (minio `PresignedGetObject`), handler больше не качает файл — `GetDuration(url)` (ffprobe тянет только moov, десятки КБ) → `seek = duration*0.1` → `GenerateThumbnail(url)` (ffmpeg делает HTTP Range-запросы, скачивает только нужный участок + moov). Проверено локально на Range-capable HTTP-сервере: ffprobe 30s, ffmpeg seek → валидный JPEG.
- **Почему:** S3/MinIO поддерживает Range; `io.ReadCloser`-стрим в stdin всё равно передаёт все байты — не решает задачу. URL короткоживущий (15 мин), виден в `ps` — приемлемо для окружения. Нюанс: python http.server (без Range) не годится как референс — с ним ffmpeg падает.
- **Примечание:** Missing source теперь определяется по ошибке ffmpeg (`404`/`NoSuchKey`/`not found`) → `asynq.SkipRetry`, т.к. presign не проверяет существование объекта.
