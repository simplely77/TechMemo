package worker

import "techmemo/backend/ai/queue"

type MemoryQueue struct {
	ch chan queue.AITask
}

func NewMemoryQueue(size int) *MemoryQueue {
	return &MemoryQueue{
		ch: make(chan queue.AITask, size),
	}
}

func (m *MemoryQueue) Publish(task queue.AITask) error {
	m.ch <- task
	return nil
}

func (m *MemoryQueue) Consume() (<-chan queue.AITask, error) {
	return m.ch, nil
}
