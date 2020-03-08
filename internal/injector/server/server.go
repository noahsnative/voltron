package server

import (
	"fmt"
	"net/http"
)

// ServerOptions represent server configuration
type ServerOptions func(s *http.Server)

// WithPort sets a TCP port a Server will be listen on
func WithPort(port int) ServerOptions {
	return func(s *http.Server) {
		s.Addr = fmt.Sprintf(":%v", port)
	}
}

// New creates an http.Server instance with provided options
func New(mux http.Handler, opts ...ServerOptions) *http.Server {
	var server http.Server
	server.Handler = mux

	for _, opt := range opts {
		opt(&server)
	}

	return &server
}
