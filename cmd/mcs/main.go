package main

import (
	"fmt"
	"os"

	"github.com/cv/mcs/internal/cli"
)

// Version is set at build time via ldflags.
var Version = "dev"

func main() {
	if err := cli.Execute(Version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
