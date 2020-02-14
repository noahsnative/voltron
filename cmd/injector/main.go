package main

import (
	"fmt"
	"os"

	"github.com/noahsnative/voltron/internal/injector"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func run() error {
	server := injector.New(injector.WithPort(8080))
	return server.Run()
}
