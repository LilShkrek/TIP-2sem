package events

const (
	TaskCreatedEvent = "task.created"
	ProducerTasks    = "tasks"
	EventVersion     = "1.0"
)

type TaskEvent struct {
	Event     string `json:"event"`
	TaskID    string `json:"task_id"`
	TS        string `json:"ts"`
	RequestID string `json:"request_id,omitempty"`
	Producer  string `json:"producer,omitempty"`
	Version   string `json:"version,omitempty"`
}
