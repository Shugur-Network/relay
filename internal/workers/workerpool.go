package workers

import (
	"sync"
	"time"
)

// WorkerPool manages a pool of workers that execute jobs concurrently.
type WorkerPool struct {
	jobCh chan func()
	wg    sync.WaitGroup
}

// NewWorkerPool initializes a worker pool with a fixed number of workers.
func NewWorkerPool(workerCount, jobBufferSize int) *WorkerPool {
	wp := &WorkerPool{
		jobCh: make(chan func(), jobBufferSize),
	}
	for i := 0; i < workerCount; i++ {
		go wp.worker()
	}
	return wp
}

// worker executes jobs and sleeps when idle to reduce CPU usage.
func (wp *WorkerPool) worker() {
	for job := range wp.jobCh {
		job()
		time.Sleep(10 * time.Millisecond) // Prevents CPU busy-waiting
	}
}

// AddJob enqueues a job without blocking.
func (wp *WorkerPool) AddJob(job func()) bool {
	wp.wg.Add(1)
	select {
	case wp.jobCh <- func() {
		defer wp.wg.Done()
		job()
	}:
		return true
	default: // Drop the job if queue is full
		wp.wg.Done()
		return false
	}
}

// Wait blocks until all jobs are completed.
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

// Stop closes the job channel gracefully.
func (wp *WorkerPool) Stop() {
	var once sync.Once
	once.Do(func() {
		close(wp.jobCh)
		wp.wg.Wait()
	})
}

// Resize changes the number of workers dynamically.
func (wp *WorkerPool) Resize(newWorkerCount int) {
	wp.Stop()
	wp.jobCh = make(chan func(), cap(wp.jobCh))
	for i := 0; i < newWorkerCount; i++ {
		go wp.worker()
	}
}
