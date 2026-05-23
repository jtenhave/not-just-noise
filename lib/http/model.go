package http

import (
	"context"
	"reflect"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type Request struct {
	Context    context.Context
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

// Code returns the code of the response.
func (r *response) Code() int {
	return r.code
}

// Body returns the body of the response.
func (r *response) Body() interface{} {
	return r.body
}

// CreateResponse creates a new Response with the given code and body.
func CreateResponse(code int, body interface{}) Response {
	return &response{
		code: code,
		body: body,
	}
}

// CreateErrorResponse creates a new Response with the given error.
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

// CreateRoute creates a new Route using the given method, path, bodyType and handler.
func CreateRoute(method string, path string, bodyType reflect.Type, handler func(Request) Response) Route {
	return Route{
		Method:   method,
		Path:     path,
		BodyType: bodyType,
		Handler:  handler,
	}
}
