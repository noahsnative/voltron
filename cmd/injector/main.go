package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/noahsnative/voltron/internal/injector/mutate"
	"github.com/noahsnative/voltron/internal/injector/server"
)

var (
	errorCode = 1
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(errorCode)
	}
}

func run() error {
	port := flag.Int("port", 8080, "TCP port to listen on.")
	certFile := flag.String("tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	keyFile := flag.String("tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		return fmt.Errorf("could not load the TLS certificate: %v", err)
	}

	mux := http.NewServeMux()
	handler := mutate.New()
	mux.HandleFunc("/mutate", handler.Mutate)
	s := server.New(mux, server.WithPort(*port), server.WithTLSCertificate(&cert))
	fmt.Printf("Listening on %d\n", *port)
	return s.ListenAndServeTLS("", "")
}
