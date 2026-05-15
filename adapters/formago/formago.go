// Package formago adapts *http.ServeMux to the forma	.Router interface.
package formago

import (
	"net/http"
)

// ServeMuxRouter adapts *http.ServeMux to htmlmux.Router.
type ServeMuxRouter struct{ mux *http.ServeMux }

// New wraps mux so it satisfies htmlmux.Router.
func New(mux *http.ServeMux) *ServeMuxRouter {
	return &ServeMuxRouter{mux: mux}
}

// Handle registers h for the given method and path pattern on the underlying ServeMux.
func (s *ServeMuxRouter) Handle(method, path string, h http.HandlerFunc) {
	s.mux.HandleFunc(method+" "+path, h)
}
