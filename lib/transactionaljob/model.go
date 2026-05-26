package transactionaljob

type TransactionalJob struct {
	ID           string
	CallbackURL  string
	Payload      string
	ClaimTimeout int
	RetrySeconds int
	RetryBackoff float64
	MaxAttempts  int
}
