package transactionaljob

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jtenhave/not-just-noise/transactional-job-service/internal/config"
)

type TransactionalJobService interface {
	GetAvailableTransactionalJobs(ctx context.Context, limit int) ([]TransactionalJob, error)
	SendTransactionalJob(ctx context.Context, transactionalJob TransactionalJob) error
}

type transactionalJobWorker struct {
	config                  config.WorkerConfig
	transactionalJobService TransactionalJobService
}

// NewTransactionalJobWorker creates a new transactionalJobWorker using the given transactionalJobService.
func NewTransactionalJobWorker(config config.WorkerConfig, transactionalJobService TransactionalJobService) transactionalJobWorker {
	return transactionalJobWorker{
		config:                  config,
		transactionalJobService: transactionalJobService,
	}
}

// Start starts the transactionalJobWorker.
func (worker *transactionalJobWorker) Start(ctx context.Context) {
	fmt.Printf("\nstarting transactional job worker\n")
	worker.work(ctx)
}

// work is the main loop for the transactionalJobWorker.
func (worker *transactionalJobWorker) work(ctx context.Context) {
	sendChannel := make(chan TransactionalJob, worker.config.JobBufferSize)
	for i := 0; i < worker.config.MaxWorkers; i++ {
		go worker.runWorker(ctx, sendChannel)
	}

	defer close(sendChannel)

	for {
		fmt.Printf("getting available transactional jobs\n")
		transactionalJobs, err := worker.transactionalJobService.GetAvailableTransactionalJobs(ctx, worker.config.MaxBatchSize)
		if err != nil {
			log.Printf("transactionaljobworker.work: failed to get available transactional jobs: %v", err)

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(worker.config.NoJobDelay) * time.Second):
			}
			continue
		}

		fmt.Printf("found %d available transactional jobs\n", len(transactionalJobs))
		for _, transactionalJob := range transactionalJobs {
			select {
			case <-ctx.Done():
				return

			case sendChannel <- transactionalJob:
			}
		}

		if len(transactionalJobs) < worker.config.MaxBatchSize {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(worker.config.NoJobDelay) * time.Second):
			}
		}
	}
}

// runWorker runs a new worker using the given sendChannel.
func (worker *transactionalJobWorker) runWorker(ctx context.Context, sendChannel <-chan TransactionalJob) {
	for {
		select {
		case <-ctx.Done():
			return
		case transactionalJob, ok := <-sendChannel:
			if !ok {
				return
			}

			fmt.Printf("sending transactional job: %s\n", transactionalJob.ID)

			err := worker.transactionalJobService.SendTransactionalJob(ctx, transactionalJob)
			if err != nil {
				log.Printf("transactionaljobworker.runWorker: failed to send transactional job: %v", err)
			}
		}
	}
}
