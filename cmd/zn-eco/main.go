package main

import (
	"os"

	"ziniao/internal/cli"
	"ziniao/internal/variant"
)

func main() {
	os.Exit(cli.Execute(variant.Eco))
}
