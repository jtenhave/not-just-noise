package transactionaljob

import "time"

type TransactionalJob struct {
	ID           string
	CallbackURL  string
	Payload      string
	CreatedAt    time.Time
	AvailableAt  time.Time
	ProcessingAt *time.Time
	Attempts     int
	LastError    *string
}
