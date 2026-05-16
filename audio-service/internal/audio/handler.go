package audio

import (
	"fmt"

	"github.com/jtenhave/not-just-noise/lib/http"
)

func CreateRoutes() []http.Route {
	getHandler := func(request http.Request) http.Response {
		return http.Response{
			StatusCode: 200,
			Body:       fmt.Sprintf("Hello, %s!", request.PathValue("id")),
		}
	}

	return []http.Route{
		http.CreateRoute("GET", "/audio/{id}", getHandler),
	}
}
