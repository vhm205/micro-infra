// Package worker provider workers for jobs
package worker

import (
	"context"
	"fmt"
	"sync"
	"worker-service/jobs"
)

type WorkerPool struct {
	ctx     context.Context
	workers int
	jobs    chan jobs.Job
	wg      sync.WaitGroup
}

func NewWorkerPool(ctx context.Context, workers int, queueSize int) *WorkerPool {
	return &WorkerPool{
		ctx:     ctx,
		workers: workers,
		jobs:    make(chan jobs.Job, queueSize),
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)

		go func(i int) {
			defer wp.wg.Done()

			for {
				select {
				case job, ok := <-wp.jobs:
					if !ok {
						return
					}
					fmt.Printf("🕙 Worker %d processing job %d\n", i, job.ID)
					job.Process(wp.ctx)
				case <-wp.ctx.Done():
					fmt.Printf("✅ Worker %d is done!\n", i)
				}
			}
		}(i)
	}

	fmt.Println("✅ Workers are started")
}

func (wp *WorkerPool) Submit(job jobs.Job) {
	select {
	case <-wp.ctx.Done():
		fmt.Printf("❌ Reject job because context cancelled\n")
		return
	default:
		wp.jobs <- job
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.jobs)
	wp.wg.Wait()
}
