package http

import (
	"reflect"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type Request struct {
	Body       interface{}
	PathValues map[string]string
}

const (
	UnknownError        = 0
	BadRequest          = 400
	NotFound            = 404
	Conflict            = 409
	InternalServerError = 500
)

type Response interface {
	Code() int
	Body() interface{}
}

type response struct {
	code int
	body interface{}
}

func (r *response) Code() int {
	return r.code
}

func (r *response) Body() interface{} {
	return r.body
}

func CreateResponse(code int, body interface{}) Response {
	return &response{
		code: code,
		body: body,
	}
}

func CreateErrorResponse(err error) Response {
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

	body := map[string]string{
		"error": err.Error(),
	}

	return &response{
		code: code,
		body: body,
	}
}

type Route struct {
	Method   string
	Path     string
	BodyType reflect.Type
	Handler  func(Request) Response
}

func CreateRoute(method string, path string, bodyType reflect.Type, handler func(Request) Response) Route {
	return Route{
		Method:   method,
		Path:     path,
		BodyType: bodyType,
		Handler:  handler,
	}
}
