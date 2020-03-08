package server

import (
	"fmt"
	"net/http"
)

// Options represent server configuration
type Options func(s *http.Server)

// WithPort sets a TCP port an http.Server will be listen on
func WithPort(port int) Options {
	return func(s *http.Server) {
		s.Addr = fmt.Sprintf(":%v", port)
	}
}

// New creates an http.Server instance with provided options
func New(mux http.Handler, opts ...Options) *http.Server {
	var server http.Server
	server.Handler = mux

	for _, opt := range opts {
		opt(&server)
	}

	return &server
}
