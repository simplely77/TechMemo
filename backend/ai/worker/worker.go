package worker

import (
	"context"
	"techmemo/backend/ai/queue"
)

type Worker struct {
	queue     queue.Queue
	handler   *Handler
	workerNum int
}

func NewWorker(queue queue.Queue, handler *Handler, workerNum int) *Worker {
	return &Worker{queue: queue, handler: handler, workerNum: workerNum}
}

func (w *Worker) Start(ctx context.Context) error {
	taskCh, err := w.queue.Consume()
	if err != nil {
		return err
	}
	for i := 0; i < w.workerNum; i++ {
		go func(id int) {
			for {
				select {
				case task := <-taskCh:
					w.handler.Handler(ctx, task)
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}
	return nil
}
