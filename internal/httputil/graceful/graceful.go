package graceful

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func Serve(
	start func() error,
	shutdown func(context.Context) error,
	timeout time.Duration,
) error {
	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		err := start()
		if err != nil && err != http.ErrServerClosed {
			serveErr <- fmt.Errorf("serve failed: %w", err)
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("server failed: %w", err)

	case <-stopCtx.Done():
		cancelCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := shutdown(cancelCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	}

	return nil
}
