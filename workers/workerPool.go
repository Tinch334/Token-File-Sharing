package workers


import (
    "context"
    "sync"

    "github.com/rs/zerolog/log"
)

// A simple worker pool that accepts jobs. Note that the jobs don't accept a response
type WorkerPool[J any, R any] struct {
    workers int
    jobs    chan J
    results chan R
    wg      sync.WaitGroup
}

// CreateWorkerPool creates a new worker pool with the specified context and values.
func CreateWorkerPool[J any, R any](workers, jobQueueSize, resultQueueSize int) *WorkerPool[J, R] {
    return &WorkerPool[J, R]{
        workers: workers,
        jobs:    make(chan J, jobQueueSize),
        results: make(chan R, resultQueueSize),
    }
}


// Starts the worker pool, with the given handler.
// The handlers get any additional information they need in the context, the handler returns (R, bool) where the boolean
// indicates if there are results to be sent to the result queue.
func (wp *WorkerPool[J, R]) StartWorkerPool(ctx context.Context, handler func(context.Context, J) (R, bool)) {
    worker := func(id int) {
        // Signal when this worker finished.
        wp.wg.Add(1)
        defer wp.wg.Done()

        for {
            select {
            // Shutdown requested by context.
            case <-ctx.Done():
                return
            case job, open := <-wp.jobs:
                // Channel is closed.
                if !open {
                    return
                }

                // Perform job and get result.
                res, hasRes := handler(ctx, job)

                // Handler gave no result.
                if !hasRes {
                    continue
                }

                // Check for a shutdown whilst waiting to send results, avoids a thread leak whilst waiting to send results.
                select{
                case <-ctx.Done():
                    return
                case wp.results <- res:
                    // Results sent.
                }
            }
        }
    }

    // Start workers.
    for i := range wp.workers {
        go worker(i)
    }

    // Start clean-up thread for results channel.
    go func() {
        wp.wg.Wait()

        log.Info().
            Msg("Worker pool clean-up")

        close(wp.results)
    }()

    log.Info().
        Msg("Started worker pool")
}


// Submit sends a job to the worker pool.
func (wp *WorkerPool[J, R]) Submit(ctx context.Context, job J) error {
    select {
    // Channel was closed before job could be submitted, return error.
    case <-ctx.Done():
        return ctx.Err()
    // Job was submitted successfully, return no error.
    case wp.jobs <- job:
        return nil
    }
}


// Results provides access to the results channel.
func (wp *WorkerPool[J, R]) Results() <-chan R{
    return wp.results
}


// Wait blocks until all workers and cleaners have finished processing.
func (wp *WorkerPool[J, R]) Wait() {
    wp.wg.Wait()
}