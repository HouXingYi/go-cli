package main

import (
	"os"

	"ziniao/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
