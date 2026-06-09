package dispatch

import (
	"context"
	"time"

	"github.com/jtenhave/not-just-noise/lib/log"
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
	log.Logger(ctx).Info("starting dispatcher")
	worker.work(ctx)
}

// work is the main loop for the dispatchWorker.
func (worker *dispatchWorker) work(ctx context.Context) {
	sendChannel := make(chan Dispatch, worker.config.JobBufferSize)
	for i := 0; i < worker.config.MaxWorkers; i++ {
		go worker.runWorker(ctx, i, sendChannel)
	}

	defer close(sendChannel)

	for {
		dispatches, err := worker.dispatchService.GetAvailableDispatches(ctx, worker.config.MaxBatchSize)
		if err != nil {
			log.Logger(ctx).Error("failed to get available dispatches", "error", err)

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(worker.config.NoJobDelay) * time.Second):
			}
			continue
		}

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
func (worker *dispatchWorker) runWorker(ctx context.Context, workerID int, sendChannel <-chan Dispatch) {
	for {
		select {
		case <-ctx.Done():
			return
		case dispatch, ok := <-sendChannel:
			if !ok {
				return
			}

			ctx = log.LoggerWithCtx(ctx, "worker_id", workerID, "dispatch_id", dispatch.ID)
			err := worker.dispatchService.Dispatch(ctx, dispatch)
			if err != nil {
				log.Logger(ctx).Error("failed to dispatch", "error", err)
				continue
			}
		}
	}
}
