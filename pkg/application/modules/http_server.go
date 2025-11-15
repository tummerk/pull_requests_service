package modules

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

type HTTPServer struct {
	ShutdownTimeout time.Duration
}

func (h HTTPServer) Run(
	gCtx context.Context,
	g *errgroup.Group,
	httpServer *http.Server,
) {
	g.Go(func() error {
		logger(gCtx).Info("http server started", slog.String("address", httpServer.Addr))

		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger(gCtx).Error("http server ListenAndServe error", slog.Any("error", err))
			return fmt.Errorf("httpServer.ListenAndServe: %w", err)
		}

		logger(gCtx).Info("http server stopped listening")
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()

		logger(gCtx).Info("http server is shutting down", slog.Duration("timeout", h.ShutdownTimeout))

		shutdownCtx, cancel := context.WithTimeout(context.Background(), h.ShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger(gCtx).Error("http server shutdown error", slog.Any("error", err))
			return err
		}

		logger(gCtx).Info("http server shut down gracefully")
		return nil
	})
}
