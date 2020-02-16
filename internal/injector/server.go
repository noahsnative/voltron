package injector

import (
	"fmt"
	"log"
	"net/http"
)

// Server represents a mutating webhook HTTP server
type Server struct {
	server *http.Server
	port   string
}

// ServerOptions represent server configuration
type ServerOptions func(s *Server)

// WithPort sets a TCP port a Server will be listen on
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

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", server.handleMutate)

	server.server.Handler = mux

	for _, opt := range opts {
		opt(&server)
	}

	return &server
}

// Run starts listening for incoming HTTP requests
func (s *Server) Run() error {
	fmt.Printf("Listening on %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

// ServerHTTP handles an HTTP request
func (s *Server) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.Handler.ServeHTTP(w, r)
}

func (s *Server) handleMutate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Invalid method %s, only POST requests are allowed", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
