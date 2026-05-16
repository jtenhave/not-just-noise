package http

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Request struct {
	Body      string
	PathValue func(string) string
}

type Response struct {
	StatusCode int
	Body       string
}

type Route struct {
	Method  string
	Path    string
	Handler func(Request) Response
}

func CreateRoute(method string, path string, handler func(Request) Response) Route {
	return Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
}

func CreateErrorResonse(code int, message string) Response {
	body := map[string]string{
		"code":  strconv.Itoa(code),
		"error": message,
	}

	errorBytes, err := json.Marshal(body)
	if err != nil {
		return Response{
			StatusCode: 500,
			Body:       fmt.Errorf("Failed to marshal error response: %w", err).Error(),
		}
	}

	return Response{
		StatusCode: code,
		Body:       string(errorBytes),
	}
}
