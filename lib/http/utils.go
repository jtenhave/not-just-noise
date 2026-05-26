package http

import "github.com/jtenhave/not-just-noise/lib/njnerror"

const (
	UnknownError        = 0
	Ok                  = 200
	NoContent           = 204
	BadRequest          = 400
	NotFound            = 404
	Conflict            = 409
	InternalServerError = 500
)

// ResponseCodeFromError returns the code for the given error.
func ResponseCodeFromError(err error) int {
	errorType := njnerror.Type(err)
	code := InternalServerError

	switch errorType {
	case njnerror.BadRequest:
		code = BadRequest
	case njnerror.NotFound:
		code = NotFound
	case njnerror.Conflict:
		code = Conflict
	}

	return code
}
