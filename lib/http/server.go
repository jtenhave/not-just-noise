package http

import (
	"fmt"
	"io"
	"net/http"
)

func StartServer(routes []Route, port int) error {
	mux := http.NewServeMux()

	fmt.Printf("registering routes:\n")
	for _, route := range routes {
		pattern := route.Method + " " + route.Path
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Errorf("Failed to read request body: %w", err).Error(), http.StatusInternalServerError)
				return
			}

			request := Request{
				Body: string(body),
				PathValue: func(name string) string {
					return r.PathValue(name)
				},
			}

			response := route.Handler(request)
			w.WriteHeader(response.StatusCode)
			w.Write([]byte(response.Body))
		})

		fmt.Printf("%s\n", pattern)
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	fmt.Printf("\nserving on port: %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
