package dispatch

import (
	"context"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
	"github.com/jtenhave/not-just-noise/lib/utils"
)

const (
	defaultMaxAttempts  = 10
	defaultRetrySeconds = 2
	defaultRetryBackoff = 1.5
)

type TxManager interface {
	WithinTx(ctx context.Context, transaction func(context.Context) error) error
}

type DispatchRepo interface {
	GetAvailableDispatches(ctx context.Context, limit int) ([]Dispatch, error)
	ClaimDispatches(ctx context.Context, ids []string) error
	ReleaseDispatch(ctx context.Context, id string, lastError string) error
	DeleteDispatch(ctx context.Context, id string) error
}

type Dispatcher interface {
	Dispatch(ctx context.Context, dispatcherType string, destination string, payload string) error
}

type dispatchService struct {
	txManager    TxManager
	dispatchRepo DispatchRepo
	dispatcher   Dispatcher
}

// NewDispatchService creates a new dispatch service using the given txManager, dispatchRepo, and dispatcher.
func NewDispatchService(txManager TxManager, dispatchRepo DispatchRepo, dispatcher Dispatcher) dispatchService {
	return dispatchService{
		txManager:    txManager,
		dispatchRepo: dispatchRepo,
		dispatcher:   dispatcher,
	}
}

// GetAvailableDispatches gets the available dispatches, up to the given limit. Returns the first available dispatches and the first error encountered.
func (service dispatchService) GetAvailableDispatches(ctx context.Context, limit int) ([]Dispatch, error) {
	var dispatches []Dispatch
	var err error
	service.txManager.WithinTx(ctx, func(ctx context.Context) error {
		dispatches, err = service.dispatchRepo.GetAvailableDispatches(ctx, limit)
		if err != nil {
			return njnerror.Wrapf("dispatchservice.GetAvailableDispatches: failed to get available dispatches: %w", err)
		}

		if len(dispatches) == 0 {
			return nil
		}

		ids := make([]string, len(dispatches))
		for i := range dispatches {
			ids[i] = dispatches[i].ID
		}

		err = service.dispatchRepo.ClaimDispatches(ctx, ids)
		if err != nil {
			return njnerror.Wrapf("dispatchservice.GetAvailableDispatches: failed to claim dispatches: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, njnerror.Wrapf("dispatchservice.GetAvailableDispatches: failed to get available dispatches: %w", err)
	}

	return dispatches, nil
}

// Dispatch dispatches the given dispatch. Returns the first error encountered.
func (service dispatchService) Dispatch(ctx context.Context, dispatch Dispatch) error {
	err := service.dispatcher.Dispatch(ctx, dispatch.CallbackType, dispatch.CallbackResource, dispatch.Payload)
	if err != nil {
		// Best effort retry to release the dispatch. Worst case, the claim expires and the dispatch can be processed again.
		err := utils.Retry(ctx, func() error {
			return service.dispatchRepo.ReleaseDispatch(ctx, dispatch.ID, err.Error())
		}, defaultMaxAttempts, defaultRetrySeconds, defaultRetryBackoff)

		if err != nil {
			return njnerror.Wrapf("dispatchservice.Dispatch: failed to release dispatch: %w", err)
		}

		return nil
	}

	// Best effort retry to delete the dispatch. Worst case, the dispatch is processed again. Downstream services will handle idempotency.
	err = utils.Retry(ctx, func() error {
		return service.dispatchRepo.DeleteDispatch(ctx, dispatch.ID)
	}, defaultMaxAttempts, defaultRetrySeconds, defaultRetryBackoff)

	if err != nil {
		return njnerror.Wrapf("dispatchservice.Dispatch: failed to delete dispatch: %w", err)
	}

	return nil
}
