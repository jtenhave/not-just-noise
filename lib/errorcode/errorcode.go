package errorcode

import "fmt"

const (
	BadRequest          = 400
	NotFound            = 404
	Conflict            = 409
	InternalServerError = 500
)

type errorCode struct {
	Code    int
	Message string
}

func (e *errorCode) Error() string {
	return e.Message
}

func NewErrorCode(code int, message string) *errorCode {
	return &errorCode{
		Code:    code,
		Message: message,
	}
}

func ErrorCode(e error) int {
	if err, ok := e.(*errorCode); ok {
		return err.Code
	}
	return 0
}

func Wrap(e error, message string) error {

	if err, ok := e.(*errorCode); ok {
		return &errorCode{
			Code:    err.Code,
			Message: fmt.Sprintf("%s: %s", message, err.Message),
		}
	}

	return fmt.Errorf("%s: %w", message, e)
}
