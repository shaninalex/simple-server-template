package httpx

import (
	"net/http"
	"slices"
)

type Middleware func(next http.Handler) http.Handler

type Mux struct {
	*http.ServeMux
	middlewares []Middleware
}

func NewServeMux() *Mux {
	return &Mux{
		ServeMux: http.NewServeMux(),
	}
}

func (s *Mux) Use(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

func (s *Mux) Handle(pattern string, handler http.Handler) {
	s.ServeMux.Handle(pattern, s.wrap(handler))
}

func (s *Mux) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.Handle(pattern, handler)
}

func (s *Mux) wrap(handler http.Handler) http.Handler {
	for _, v := range slices.Backward(s.middlewares) {
		handler = v(handler)
	}

	return handler
}
