package queue

import "github.com/google/uuid"

const TaskThumbsnailProcessor = "video:thumbsnail"

type ThumbsnailProcessorPayload struct {
	StreamUUID uuid.UUID `json:"stream_uuid"`
	InputPath  string    `json:"input_path"`
}
