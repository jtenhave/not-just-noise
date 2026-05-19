package errorcode

const (
	BadRequest = 400
	NotFound = 404
	Conflict = 409
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


