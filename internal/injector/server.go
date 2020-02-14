package injector

import (
	"fmt"
	"net/http"
)

// Server represents a mutating webhook HTTP server
type Server struct {
	server *http.Server
	port   string
}

// ServerOptions represent server configuration
type ServerOptions func(s *Server)

// WithPort sets a port a Server is going to listen
func WithPort(port int) ServerOptions {
	return func(s *Server) {
		s.server.Addr = fmt.Sprintf(":%v", port)
	}
}

// New creates a Server instance with provided options
func New(opts ...ServerOptions) *Server {
	server := Server{
		server: &http.Server{},
	}

	for _, opt := range opts {
		opt(&server)
	}

	return &server
}

// Run starts listening for incoming HTTP requests
func (s *Server) Run() error {
	return s.server.ListenAndServe()
}
