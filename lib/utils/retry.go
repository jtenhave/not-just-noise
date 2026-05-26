package utils

import (
	"context"
	"time"
)

func Retry(ctx context.Context, fn func() error, maxAttempts int, delaySeconds int64, delayBackoff float64) error {
	var lastError error
	nextDelay := delaySeconds
	for i := 0; i < maxAttempts; i++ {
		lastError := fn()
		if lastError == nil {
			return nil
		}

		time.Sleep(time.Duration(nextDelay) * time.Second)
		nextDelay = int64(float64(nextDelay) * delayBackoff)
	}

	return lastError
}
