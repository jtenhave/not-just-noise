package dispatch

import (
	"context"
	"fmt"
	"log"
	"time"
)

type WorkerConfig struct {
	MaxWorkers    int
	MaxBatchSize  int
	NoJobDelay    int
	JobBufferSize int
}

type DispatchService interface {
	GetAvailableDispatches(ctx context.Context, limit int) ([]Dispatch, error)
	Dispatch(ctx context.Context, dispatch Dispatch) error
}

type dispatchWorker struct {
	config          WorkerConfig
	dispatchService DispatchService
}

// NewDispatchWorker creates a new dispatchWorker using the given dispatchService.
func NewDispatchWorker(config WorkerConfig, dispatchService DispatchService) dispatchWorker {
	return dispatchWorker{
		config:          config,
		dispatchService: dispatchService,
	}
}

// Start starts the dispatchWorker.
func (worker *dispatchWorker) Start(ctx context.Context) {
	fmt.Printf("\nstarting transactional job worker\n")
	worker.work(ctx)
}

// work is the main loop for the dispatchWorker.
func (worker *dispatchWorker) work(ctx context.Context) {
	sendChannel := make(chan Dispatch, worker.config.JobBufferSize)
	for i := 0; i < worker.config.MaxWorkers; i++ {
		go worker.runWorker(ctx, sendChannel)
	}

	defer close(sendChannel)

	for {
		fmt.Printf("getting available dispatches\n")
		dispatches, err := worker.dispatchService.GetAvailableDispatches(ctx, worker.config.MaxBatchSize)
		if err != nil {
			log.Printf("dispatchworker.work: failed to get available dispatches: %v", err)

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(worker.config.NoJobDelay) * time.Second):
			}
			continue
		}

		fmt.Printf("found %d available dispatches\n", len(dispatches))
		for _, dispatch := range dispatches {
			select {
			case <-ctx.Done():
				return

			case sendChannel <- dispatch:
			}
		}

		if len(dispatches) < worker.config.MaxBatchSize {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(worker.config.NoJobDelay) * time.Second):
			}
		}
	}
}

// runWorker runs a new worker using the given sendChannel.
func (worker *dispatchWorker) runWorker(ctx context.Context, sendChannel <-chan Dispatch) {
	for {
		select {
		case <-ctx.Done():
			return
		case dispatch, ok := <-sendChannel:
			if !ok {
				return
			}

			fmt.Printf("sending dispatch: %s\n", dispatch.ID)

			err := worker.dispatchService.Dispatch(ctx, dispatch)
			if err != nil {
				log.Printf("dispatchworker.runWorker: failed to dispatch: %v", err)
			}
		}
	}
}
