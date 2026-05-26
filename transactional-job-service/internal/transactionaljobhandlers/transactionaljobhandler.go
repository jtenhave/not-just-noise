package transactionaljobhandlers

import "context"

type TransactionalJobHandler interface {
	Handle(ctx context.Context, destination string, payload string) error
}