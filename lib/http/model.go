package http

import (
	"context"
)

type Request struct {
	Context    context.Context
	Body       string
	PathValues map[string]string
}

type Response struct {
	Code    int
	Body    *string
	Headers map[string]string
}

type Route struct {
	Method  string
	Path    string
	Handler func(Request) Response
}
