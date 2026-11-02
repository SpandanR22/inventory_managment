package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SpandanR22/inventory_managment/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// run returns instead of exiting, so its deferred calls (closing the
	// pool) always execute. os.Exit is only called here, at the very end.
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// A missing .env is fine (production sets real environment variables).
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("loading .env: %w", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL environment variable is not set")
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8086"
	}

	// ctx is cancelled when the process receives Ctrl+C or SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()
	slog.Info("connected to database")

	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      routes(pool),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return serve(ctx, srv)
}

// serve runs the server until ctx is cancelled, then shuts down gracefully:
// it stops accepting connections and waits for in-flight requests to finish.
func serve(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("starting server", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
