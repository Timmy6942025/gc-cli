package main

import (
	"os"

	"github.com/timothy/gc-cli/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
