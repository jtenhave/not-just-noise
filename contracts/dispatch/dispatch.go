package dispatch

type Dispatch struct {
	ID               string
	CallbackType     string
	CallbackResource string
	Payload          string
	ClaimTimeout     int
	RetrySeconds     int
	RetryBackoff     float64
	MaxAttempts      int
}
