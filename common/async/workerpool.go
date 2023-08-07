package async

import (
	"go.uber.org/zap"
)

type Job interface {
	Do()
}

type WorkerPool struct {
	workerCount int
	jobQueue    chan Job
	workerQueue chan chan Job
	workers     []*worker
	quit        chan struct{}
	log         *zap.Logger
}

func NewWorkerPool(maxWorkers, maxQueue int, log *zap.Logger) *WorkerPool {
	return &WorkerPool{
		workerCount: maxWorkers,
		jobQueue:    make(chan Job, maxQueue),
		workerQueue: make(chan chan Job, maxWorkers),
		quit:        make(chan struct{}),
		log:         log,
	}
}

func (w *WorkerPool) Run() *WorkerPool {
	for i := 0; i < w.workerCount; i++ {
		worker := NewWorker(i, w.log)
		worker.Run(w.workerQueue)
		w.workers = append(w.workers, worker)
	}
	go func() {
		for {
			select {
			case job := <-w.jobQueue:
				worker := <-w.workerQueue
				worker <- job
			case <-w.quit:
				for _, worker := range w.workers {
					worker.Close()
				}
				return
			}
		}
	}()
	return w
}

func (w *WorkerPool) WorkerCount() int {
	return w.workerCount
}

func (w *WorkerPool) Add(job Job) {
	w.jobQueue <- job
}

func (w *WorkerPool) Close() {
	go func() {
		w.quit <- struct{}{}
	}()
}
