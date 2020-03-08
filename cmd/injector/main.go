package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/noahsnative/voltron/internal/injector/mutate"
	"github.com/noahsnative/voltron/internal/injector/server"
)

var (
	port = flag.Int("port", 8080, "TCP port to listen on.")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() error {
	mux := http.NewServeMux()
	handler := mutate.New()
	mux.HandleFunc("/mutate", handler.Mutate)
	s := server.New(mux, server.WithPort(*port))
	fmt.Printf("Listening on %d\n", port)
	return s.ListenAndServe()
}
