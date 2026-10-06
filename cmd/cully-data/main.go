package main

import (
	"fmt"
	"github.com/mcp-runtime/cully/internal/app"
	"os"
)

func main() {
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	var err error
	switch command {
	case "health":
		err = app.Health()
	case "serve", "migrate", "reindex":
		err = app.Data(command)
	default:
		err = fmt.Errorf("usage: cully-data [serve|migrate|reindex|health]")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
