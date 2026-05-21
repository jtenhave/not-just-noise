package njnerror

import "fmt"

type ErrorType int

const (
	UnknownError        ErrorType = 0
	BadRequest          ErrorType = 1
	NotFound            ErrorType = 2
	Conflict            ErrorType = 3
	InternalServerError ErrorType = 4
)

type njnerror struct {
	ErrorType ErrorType
	Message   string
}

func (e *njnerror) Error() string {
	return e.Message
}

func NewNJNError(errorType ErrorType, message string) error {
	return &njnerror{
		ErrorType: errorType,
		Message:   message,
	}
}

func Type(e error) ErrorType {
	if err, ok := e.(*njnerror); ok {
		return err.ErrorType
	}
	return UnknownError
}

func Wrapf(message string, args ...interface{}) error {
	errorType := UnknownError
	for _, arg := range args {
		if err, ok := arg.(*njnerror); ok {
			errorType = err.ErrorType
			break
		}
	}

	wrappedError := fmt.Errorf(message, args...)
	if errorType == UnknownError {
		return wrappedError
	}

	return &njnerror{
		ErrorType: errorType,
		Message:   wrappedError.Error(),
	}
}
