package main

import (
	"fmt"
	"github.com/mcp-runtime/cully/internal/app"
	"os"
)

var version = "dev"

func main() {
	if err := app.MCP(version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
