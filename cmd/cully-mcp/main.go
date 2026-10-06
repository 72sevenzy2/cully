package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mcp-runtime/cully/internal/app"
)

var version = "dev"

func main() {
	oauth := flag.Bool("oauth", false, "enable OAuth bearer validation (requires issuer, resource and JWKS configuration)")
	flag.Parse()
	if len(flag.Args()) != 0 {
		fmt.Fprintln(os.Stderr, "cully-mcp: unexpected arguments")
		os.Exit(2)
	}
	if err := app.MCP(version, *oauth); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
