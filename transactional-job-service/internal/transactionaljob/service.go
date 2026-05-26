package transactionaljob

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jtenhave/not-just-noise/lib/http"
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

type HTTPClient interface {
	Post(ctx context.Context, url string, body interface{}) (http.Response, error)
}

type transactionalJobService struct {
	transactionManager    TransactionManager
	transactionalJobsRepo TransactionalJobsRepo
	httpClient            HTTPClient
}

type CallbackResponseBody struct {
	Error string `json:"error"`
}

// NewTransactionalJobService creates a new transactionalJobService using the given transactionManager, transactionalJobsRepo, and httpClient.
func NewTransactionalJobService(transactionManager TransactionManager, transactionalJobsRepo TransactionalJobsRepo, httpClient HTTPClient) transactionalJobService {
	return transactionalJobService{
		transactionManager:    transactionManager,
		transactionalJobsRepo: transactionalJobsRepo,
		httpClient:            httpClient,
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
	response, err := transactionalJobService.httpClient.Post(ctx, transactionalJob.CallbackURL, transactionalJob.Payload)
	var lastErrorMessage *string
	if err != nil {
		errorMessage := err.Error()
		lastErrorMessage = &errorMessage
	} else if response.Code != http.Ok && response.Code != http.NoContent {
		errorMessage := "unknown error"
		if response.Body != nil {
			var callbackResponseBody CallbackResponseBody
			err = json.Unmarshal([]byte(*response.Body), &callbackResponseBody)
			if err != nil {
				errorMessage = *response.Body
			}
	
			errorMessage = callbackResponseBody.Error
		}

		lastErrorMessage = &errorMessage
	}

	if lastErrorMessage != nil { 
		errorMessage := fmt.Sprintf("callback returned code: %d, message: %s", response.Code, *lastErrorMessage)
		
		// Best effort retry. Worst case, the claim expires and the job can be processed again.
		err = utils.Retry(ctx, func() error {
			return transactionalJobService.transactionalJobsRepo.ReleaseTransactionalJob(ctx, transactionalJob.ID, errorMessage)
		}, defaultMaxAttempts, defaultRetrySeconds, defaultRetryBackoff)

		if err != nil {
			return njnerror.Wrapf("transactionaljobservice.SendTransactionalJob: failed to release transactional job: %w", err)
		}

		return nil
	}

	// Best effort retry. Worst case, the job is processed again. Downstream services will handle idempotency.
	err = utils.Retry(ctx, func() error {
		return transactionalJobService.transactionalJobsRepo.DeleteTransactionalJob(ctx, transactionalJob.ID)
	}, defaultMaxAttempts, defaultRetrySeconds, defaultRetryBackoff)

	if err != nil {
		return njnerror.Wrapf("transactionaljobservice.SendTransactionalJob: failed to delete transactional job: %w", err)
	}

	return nil
}
