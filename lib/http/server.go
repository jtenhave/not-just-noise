package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

// StartServer starts the server using the given routes and port.
func StartServer(routes []Route, port int) error {
	mux := http.NewServeMux()

	routes = append(routes, Route{
		Method: "GET",
		Path:   "/health",
		Handler: func(request Request) Response {
			return CreateResponse(204, nil)
		},
	})

	fmt.Printf("registering routes:\n")
	for _, route := range routes {
		pattern := route.Method + " " + route.Path
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			body, err := deserializeBody(route.BodyType, r.Body)
			if err != nil {
				sendResponse(CreateErrorResponse(err), w)
				return
			}

			request := Request{
				Context:    r.Context(),
				PathValues: extractPathValues(route.Path, r),
				Body:       body,
			}

			response := route.Handler(request)
			sendResponse(response, w)
		})

		fmt.Printf("%s\n", pattern)
	}

	fmt.Printf("\nserving on port: %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

// extractPathValues extracts values from a path template using a given request
func extractPathValues(path string, request *http.Request) map[string]string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(path, -1)

	routePathValues := make(map[string]string)
	for _, match := range matches {
		routePathValues[match[1]] = request.PathValue(match[1])
	}

	return routePathValues
}

// deserializeBody deserializes a request body into a given bodyType
func deserializeBody(bodyType reflect.Type, body io.Reader) (interface{}, error) {
	if bodyType == nil {
		return nil, nil
	}

	// Read the request body into a byte slice
	rawBody, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("libhttp.deserializeBody: failed to read request body: %w", err)
	}

	// Create a pointer to a new instance of the body type
	bodyPointer := reflect.New(bodyType).Interface()

	// Unmarshal the request body into the body pointer
	err = json.Unmarshal(rawBody, bodyPointer)
	if err != nil {
		wrappedError := fmt.Errorf("libhttp.deserializeBody: failed to unmarshal request body: %w", err)
		return nil, njnerror.NewNJNError(njnerror.BadRequest, wrappedError.Error())
	}

	// Dereference the body pointer to get the body value
	bodyPointerValue := reflect.ValueOf(bodyPointer)
	if bodyPointerValue.Kind() == reflect.Ptr {
		return bodyPointerValue.Elem().Interface(), nil
	} else {
		return nil, fmt.Errorf("libhttp.deserializeBody: failed to dereference body pointer: %w", err)
	}
}

func sendResponse(response Response, w http.ResponseWriter) {
	if response.Body() != nil {
		body, err := json.Marshal(response.Body())
		if err != nil {
			http.Error(w, fmt.Errorf("libhttp.sendResponse: failed to marshal response body: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.Code())
		w.Write(body)
	} else {
		w.WriteHeader(response.Code())
	}
}
