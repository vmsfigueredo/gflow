package main

import (
	"fmt"
	"os"

	"github.com/vmsfigueredo/gflow/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
