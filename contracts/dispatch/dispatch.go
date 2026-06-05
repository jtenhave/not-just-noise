package dispatch

type CallbackType string

const (
	CallbackTypeLog    CallbackType = "log"
	CallbackTypeNotify CallbackType = "notify"
	CallbackTypeQueue  CallbackType = "queue"
)

type Dispatch struct {
	ID               string
	CallbackType     CallbackType
	CallbackResource string
	Payload          string
	ClaimTimeout     *int
	RetrySeconds     *int
	RetryBackoff     *float64
	MaxAttempts      *int
}
