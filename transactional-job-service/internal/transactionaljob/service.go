package transactionaljob

import (
	"context"
	"fmt"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
	"github.com/jtenhave/not-just-noise/lib/utils"
)

const (
	defaultMaxAttempts  = 10
	defaultRetrySeconds = 2
	defaultRetryBackoff = 1.5
)

type TransactionManager interface {
	WithinTx(ctx context.Context, transaction func(context.Context) error) error
}

type TransactionalJobsRepo interface {
	GetAvailableTransactionalJobs(ctx context.Context, limit int) ([]TransactionalJob, error)
	ClaimTransactionalJobs(ctx context.Context, ids []string) error
	ReleaseTransactionalJob(ctx context.Context, id string, lastError string) error
	DeleteTransactionalJob(ctx context.Context, id string) error
}

type TransactionalJobHandler interface {
	Handle(ctx context.Context, destination string, payload string) error
}

type transactionalJobService struct {
	transactionManager    TransactionManager
	transactionalJobsRepo TransactionalJobsRepo
	handlers              map[string]TransactionalJobHandler
}

// NewTransactionalJobService creates a new transactionalJobService using the given transactionManager, transactionalJobsRepo, and handlers.
func NewTransactionalJobService(transactionManager TransactionManager, transactionalJobsRepo TransactionalJobsRepo, handlers map[string]TransactionalJobHandler) transactionalJobService {
	return transactionalJobService{
		transactionManager:    transactionManager,
		transactionalJobsRepo: transactionalJobsRepo,
		handlers:              handlers,
	}
}

// GetAvailableTransactionalJobs gets the next available transactional jobs, up to the given limit. Returns the first available transactional jobs and the first error encountered.
func (transactionalJobService transactionalJobService) GetAvailableTransactionalJobs(ctx context.Context, limit int) ([]TransactionalJob, error) {
	var transactionalJobs []TransactionalJob
	var err error
	transactionalJobService.transactionManager.WithinTx(ctx, func(ctx context.Context) error {
		transactionalJobs, err = transactionalJobService.transactionalJobsRepo.GetAvailableTransactionalJobs(ctx, limit)
		if err != nil {
			return njnerror.Wrapf("transactionaljobservice.GetAvailableTransactionalJobs: failed to get available transactional jobs: %w", err)
		}

		if len(transactionalJobs) == 0 {
			return nil
		}

		ids := make([]string, len(transactionalJobs))
		for i := range transactionalJobs {
			ids[i] = transactionalJobs[i].ID
		}

		err = transactionalJobService.transactionalJobsRepo.ClaimTransactionalJobs(ctx, ids)
		if err != nil {
			return njnerror.Wrapf("transactionaljobservice.GetAvailableTransactionalJobs: failed to claim transactional jobs: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, njnerror.Wrapf("transactionaljobservice.GetNextTransactionalJob: failed to get next transactional job: %w", err)
	}

	return transactionalJobs, nil
}

// SendTransactionalJob sends the given transactionalJob. Returns the first error encountered.
func (transactionalJobService transactionalJobService) SendTransactionalJob(ctx context.Context, transactionalJob TransactionalJob) error {
	var lastErrorMessage *string
	handler, ok := transactionalJobService.handlers[transactionalJob.CallbackType]
	if ok {
		err := handler.Handle(ctx, transactionalJob.CallbackResource, transactionalJob.Payload)
		if err != nil {
			handlerErrorMessage := err.Error()
			lastErrorMessage = &handlerErrorMessage
		}
	} else {
		noHandlerError := fmt.Sprintf("transactionaljobservice.SendTransactionalJob no handler found for callback type: %s", transactionalJob.CallbackType)
		lastErrorMessage = &noHandlerError
	}

	if lastErrorMessage != nil {
		// Best effort retry. Worst case, the claim expires and the job can be processed again.
		err := utils.Retry(ctx, func() error {
			return transactionalJobService.transactionalJobsRepo.ReleaseTransactionalJob(ctx, transactionalJob.ID, *lastErrorMessage)
		}, defaultMaxAttempts, defaultRetrySeconds, defaultRetryBackoff)

		if err != nil {
			return njnerror.Wrapf("transactionaljobservice.SendTransactionalJob: failed to release transactional job: %w", err)
		}

		return nil
	}

	// Best effort retry. Worst case, the job is processed again. Downstream services will handle idempotency.
	err := utils.Retry(ctx, func() error {
		return transactionalJobService.transactionalJobsRepo.DeleteTransactionalJob(ctx, transactionalJob.ID)
	}, defaultMaxAttempts, defaultRetrySeconds, defaultRetryBackoff)

	if err != nil {
		return njnerror.Wrapf("transactionaljobservice.SendTransactionalJob: failed to delete transactional job: %w", err)
	}

	return nil
}
