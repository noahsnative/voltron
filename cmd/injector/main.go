package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/noahsnative/voltron/internal/injector"
)

var (
	port = flag.Int("port", 8080, "TCP port to listen on.")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() error {
	flag.Parse()

	server := injector.New(injector.WithPort(*port))

	return server.Run()
}
