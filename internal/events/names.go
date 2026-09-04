package events

import "github.com/OshadhaVimuB/Leneki/internal/apperr"

const (
	JobState     = "job:state"
	JobProgress  = "job:progress"
	JobSegments  = "job:segments"
	JobDone      = "job:done"
	QueueChanged = "queue:changed"

	ModelProgress = "model:progress"
	ModelDone     = "model:done"
	ModelError    = "model:error"

	BenchProgress = "bench:progress"
	BenchDone     = "bench:done"
)

type JobStatePayload struct {
	ID    string        `json:"id"`
	State string        `json:"state"`
	Error *apperr.Error `json:"error,omitempty"`
}

type JobProgressPayload struct {
	ID         string  `json:"id"`
	Percent    float64 `json:"percent"`
	ETASeconds int     `json:"etaSeconds"`
}

type JobDonePayload struct {
	ID string `json:"id"`
}

type ModelProgressPayload struct {
	Name       string `json:"name"`
	BytesDone  int64  `json:"bytesDone"`
	BytesTotal int64  `json:"bytesTotal"`
}

type ModelDonePayload struct {
	Name string `json:"name"`
}

type ModelErrorPayload struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BenchProgressPayload struct {
	Stage string `json:"stage"`
}
