package dispatcher

import (
	"context"
	"fmt"

	"github.com/jtenhave/not-just-noise/lib/log"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

const (
	DispatcherTypeLog    = "log"
	DispatcherTypeNotify = "notify"
	DispatcherTypeQueue  = "queue"
)

type DispatcherHandler interface {
	Dispatch(ctx context.Context, destination string, payload string) error
}

type Dispatcher interface {
	RegisterDispatcher(dispatcherType string, dispatcherHandler DispatcherHandler)
	Dispatch(ctx context.Context, dispatcherType string, destination string, payload string) error
}

type dispatcher struct {
	handlers map[string]DispatcherHandler
}

func NewDispatcher() Dispatcher {
	return &dispatcher{
		handlers: make(map[string]DispatcherHandler),
	}
}

func (dispatcher *dispatcher) RegisterDispatcher(dispatcherType string, dispatcherHandler DispatcherHandler) {
	dispatcher.handlers[dispatcherType] = dispatcherHandler
}

func (dispatcher *dispatcher) Dispatch(ctx context.Context, dispatcherType string, destination string, payload string) error {
	if dispatcher.handlers[dispatcherType] == nil {
		return fmt.Errorf("dispatcher.Dispatch: handler not found for dispatcher type: %s", dispatcherType)
	}

	log.Logger(ctx).Info("dispatching message", "dispatcher_type", dispatcherType, "destination", destination)
	
	err := dispatcher.handlers[dispatcherType].Dispatch(ctx, destination, payload)
	if err != nil {
		return njnerror.Wrapf("dispatcher.Dispatch: failed to dispatch message: %w", err)
	}

	return nil
}
