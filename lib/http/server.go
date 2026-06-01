package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

const (
	UnknownError        = 0
	Ok                  = 200
	NoContent           = 204
	BadRequest          = 400
	NotFound            = 404
	Conflict            = 409
	InternalServerError = 500
)

type Route struct {
	Method  string
	Path    string
	Handler func(request *http.Request, response http.ResponseWriter)
}

// StartServer starts the server using the given routes and port.
func StartServer(routes []Route, port int) error {
	mux := http.NewServeMux()

	routes = append(routes, Route{
		Method: "GET",
		Path:   "/health",
		Handler: func(request *http.Request, response http.ResponseWriter) {
			response.WriteHeader(204)
		},
	})

	fmt.Printf("registering routes:\n")
	for _, route := range routes {
		pattern := route.Method + " " + route.Path
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			route.Handler(r, w)
		})

		fmt.Printf("%s\n", pattern)
	}

	fmt.Printf("\nserving on port: %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func responseCodeFromError(err error) int {
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

func SendErrorResponse(w http.ResponseWriter, err error) {
	code := responseCodeFromError(err)
	body := map[string]string{
		"error": err.Error(),
	}

	SendJsonResponse(w, code, &body)
}

func SendJsonResponse(w http.ResponseWriter, code int, body interface{}) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		log.Printf("libhttp.server.SendJsonResponse: failed to marshal body: %v, error: %v", body, err)
		SendResponse(w, code, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	SendResponse(w, code, bodyBytes)
}

func SendResponse(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	if body != nil {
		_, err := w.Write(body)
		if err != nil {
			log.Printf("libhttp.server.SendResponse: failed to write body: %v, error: %v\n", string(body), err)
		}
	}
}

func ReadAllAndUnmarshal(r io.Reader, v interface{}) error {
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("libhttp.server.ReadAllAndUnmarshal: failed to read all bytes from request: %v", err)
	}

	err = json.Unmarshal(bodyBytes, v)
	if err != nil {
		return fmt.Errorf("libhttp.server.ReadAllAndUnmarshal: failed to unmarshal: %v", err)
	}

	return nil
}