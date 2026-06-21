package jobs

import "errors"

const ProcessTaskJob = "process_task"

var ErrEmptyMessageID = errors.New("message_id is required")

type TaskJob struct {
	Job       string `json:"job"`
	TaskID    string `json:"task_id"`
	Attempt   int    `json:"attempt"`
	MessageID string `json:"message_id"`
}

func (j TaskJob) ValidateMessageID() error {
	if j.MessageID == "" {
		return ErrEmptyMessageID
	}
	return nil
}
