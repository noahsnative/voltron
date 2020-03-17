package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// Option represent server configuration
type Option func(s *http.Server)

// WithPort sets a TCP port an http.Server will be listen on
func WithPort(port int) Option {
	return func(s *http.Server) {
		s.Addr = fmt.Sprintf(":%v", port)
	}
}

// WithTLSCertificate adds a TLS cerificate/private key for HTTPS
func WithTLSCertificate(cert *tls.Certificate) Option {
	return func(s *http.Server) {
		s.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{*cert},
		}
	}
}

// New creates an http.Server instance with provided options
func New(mux http.Handler, opts ...Option) *http.Server {
	var server http.Server
	server.Handler = mux

	for _, opt := range opts {
		opt(&server)
	}

	return &server
}
