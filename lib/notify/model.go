package notify

import "time"

type TransactionalOutboxRecord struct {
	ID        string
	Payload   string
	CreatedAt time.Time
}
