package injectz

import (
	"net/http"
)

// NewMiddleware creates a new HTTP server middleware that enriches context using the given Injector.
func NewMiddleware(injector Injector) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(injector(r.Context())))
		})
	}
}
