package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/cv/mcs/internal/cli"
)

// Version is set at build time via ldflags.
var Version = "dev"

func main() {
	if err := cli.Execute(Version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if e, ok := errors.AsType[interface {
			error
			ExitCode() int
		}](err); ok {
			os.Exit(e.ExitCode())
		}
		os.Exit(1)
	}
}
