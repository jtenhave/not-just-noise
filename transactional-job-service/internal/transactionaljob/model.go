package transactionaljob

import "time"

type TransactionalJob struct {
	ID               string
	CallbackType     string
	CallbackResource string
	Payload          string
	CreatedAt        time.Time
	AvailableAt      time.Time
	ProcessingAt     *time.Time
	Attempts         int
	LastError        *string
}
