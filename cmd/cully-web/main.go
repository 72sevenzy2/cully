package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcp-runtime/cully/internal/website"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func run() error {
	s, err := website.New(env("CULLY_WEB_SITE_DIR", "/srv/site"), env("CULLY_WEB_DATA_DIR", "/var/lib/cully-web"))
	if err != nil {
		return err
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			entries, err := website.List(s.DataDir, "pending")
			if err != nil {
				return err
			}
			return json.NewEncoder(os.Stdout).Encode(entries)
		case "approve", "reject":
			if len(os.Args) != 3 {
				return fmt.Errorf("usage: cully-web %s ID", os.Args[1])
			}
			return website.Review(s.DataDir, os.Args[1], os.Args[2])
		default:
			return fmt.Errorf("usage: cully-web [list|approve ID|reject ID]")
		}
	}
	server := &http.Server{Addr: env("CULLY_WEB_LISTEN", ":8080"), Handler: s.Handler(),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()
	log.Printf("Cully website listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
