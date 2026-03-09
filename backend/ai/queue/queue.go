package queue

type Queue interface {
	Publish(task AITask) error
	Consume() (<-chan AITask, error)
}

type AITask struct {
	TaskID string
}