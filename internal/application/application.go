package application

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"pull_requests_service/internal/config"
	"pull_requests_service/internal/domain/service"
	"pull_requests_service/internal/infrastructure/persistence"
	"pull_requests_service/internal/server"
	"pull_requests_service/pkg/application/connectors"
	"pull_requests_service/pkg/application/modules"
	"pull_requests_service/pkg/contextx"
	"pull_requests_service/pkg/middlewarex"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/samber/lo"
)

var logger = contextx.LoggerFromContextOrDefault //nolint:gochecknoglobals

type App struct {
	cfg        config.Config
	slog       *connectors.Slog
	postgres   *connectors.Postgres
	httpServer modules.HTTPServer

	userRepo *persistence.UserRepository
	teamRepo *persistence.TeamRepository
	prRepo   *persistence.PullRequestRepository

	userService *service.UserService
	teamService *service.UserService
	prService   *service.PullRequestService
}

func New(appVersion string) App {
	const appName = "pr_service"

	cfg := lo.Must(config.Load())

	return App{
		cfg: cfg,
		slog: &connectors.Slog{
			Name:    appName,
			Version: appVersion,
			Debug:   cfg.Debug,
		},
		postgres: &connectors.Postgres{
			DSN:             cfg.Postgres.DSN,
			MaxIdleConns:    cfg.Postgres.MaxIdleConns,
			MaxOpenConns:    cfg.Postgres.MaxOpenConns,
			ConnMaxLifetime: cfg.Postgres.ConnMaxLifetime,
		},

		httpServer: modules.HTTPServer{
			ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
		},
	}
}

func (app App) shutdown(ctx context.Context) {
	app.postgres.Close(ctx)
}

func (app App) Run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer stop()

	ctx = contextx.WithLogger(ctx, app.slog.Logger(ctx))

	defer app.shutdown(ctx)

	logger(ctx).Info("config", slog.Any("config", app.cfg))

	client := app.postgres.Client(ctx)
	app.userRepo = persistence.NewUserRepository(client)
	app.teamRepo = persistence.NewTeamRepository(client)
	app.prRepo = persistence.NewPullRequestRepository(client)

	app.userService = service.NewUserService(app.userRepo)
	g, ctx := errgroup.WithContext(ctx)
	if err := g.Wait(); err != nil {
		return fmt.Errorf("g.Wait: %w", err)
	}

	return nil
}

func (app App) newHTTPServer(ctx context.Context, exampleService example.Service) *http.Server { //nolint:funlen,maintidx
	router := chi.NewRouter()

	router.Use(
		middleware.RealIP,
		middlewarex.Logger,
	)

	server.NewServer(
		server.NewExampleServer(exampleService),
	).RegisterRoutes(router)

	return &http.Server{
		//nolint:exhaustruct
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Addr:              app.cfg.HTTP.ListenAddress,
		WriteTimeout:      app.cfg.HTTP.WriteTimeout,
		ReadTimeout:       app.cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: app.cfg.HTTP.ReadTimeout,
		IdleTimeout:       app.cfg.HTTP.IdleTimeout,
		Handler:           router,
	}
}
