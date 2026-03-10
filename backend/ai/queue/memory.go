package queue

type MemoryQueue struct {
	ch chan AITask
}

func NewMemoryQueue(size int) *MemoryQueue {
	return &MemoryQueue{
		ch: make(chan AITask, size),
	}
}

func (m *MemoryQueue) Publish(task AITask) error {
	m.ch <- task
	return nil
}

func (m *MemoryQueue) Consume() (<-chan AITask, error) {
	return m.ch, nil
}
