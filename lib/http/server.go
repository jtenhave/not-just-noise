package http

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// StartServer starts the server using the given routes and port.
func StartServer(routes []Route, port int) error {
	mux := http.NewServeMux()

	routes = append(routes, Route{
		Method: "GET",
		Path:   "/health",
		Handler: func(request Request) Response {
			return Response{
				Code: 204,
				Body: nil,
			}
		},
	})

	fmt.Printf("registering routes:\n")
	for _, route := range routes {
		pattern := route.Method + " " + route.Path
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			body, err := deserializeBody(r.Body)
			if err != nil {
				panic(err)
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
func deserializeBody(body io.Reader) (string, error) {
	rawBody, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("libhttp.deserializeBody: failed to read request body: %w", err)
	}

	return string(rawBody), nil
}

func sendResponse(response Response, w http.ResponseWriter) {
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	if response.Body != nil {
		w.WriteHeader(response.Code)
		w.Write([]byte(*response.Body))
	} else {
		w.WriteHeader(response.Code)
	}
}
