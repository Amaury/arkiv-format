package main

import (
	"fmt"
	"os"

	arkivformat "github.com/Amaury/arkiv-format/go/internal/arkiv-format"
)

// main is the entrypoint. It delegates argument parsing and command handling
// to the arkivformat package.
func main() {
	if err := arkivformat.RunCLI(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

