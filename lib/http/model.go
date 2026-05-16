package http

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
