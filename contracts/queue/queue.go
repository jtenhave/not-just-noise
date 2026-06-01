package queue

import "context"

type Message struct {
	Body   *string
	Delete func(ctx context.Context) error
}
