package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/phenirain/sso/internal/application"
	"github.com/phenirain/sso/internal/config"
	"github.com/phenirain/sso/internal/lib/jwt"
	"github.com/phenirain/sso/pkg/database"
	"github.com/phenirain/sso/pkg/logger"
	"golang.org/x/sync/errgroup"
)

func Run(cfg *config.Config) error {

	if err := logger.Setup(cfg.Env); err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	db := database.MustInitDb(cfg.ConnectionString)

	g, ctx := errgroup.WithContext(context.Background())
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	startServers(ctx, g, db, cfg)
	startPprofServer(ctx, g)

	if err := g.Wait(); err != nil && errors.Is(err, context.Canceled) {
		return fmt.Errorf("server exited with error: %w", err)
	}

	return nil
}

func startServers(ctx context.Context, g *errgroup.Group, db *sqlx.DB, cfg *config.Config) {
	jwtLib := jwt.NewJwtLib(time.Minute*60, []byte(cfg.Secret))

	log := slog.Default()

	httpServer, err := application.SetupHTTPServer(cfg, db, jwtLib, log)
	if err != nil {
		slog.Error("Failed to setup HTTP server", "err", err)
		return
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           httpServer,
		ReadHeaderTimeout: time.Second * 5,
	}

	startGroup(ctx, g, "http", fmt.Sprintf("%d", cfg.HTTP.Port), server, time.Second*5)
}

func startPprofServer(ctx context.Context, g *errgroup.Group) {
	pprofAddress := fmt.Sprintf("0.0.0.0:%d", 6060)
	pprofServer := &http.Server{Addr: pprofAddress, Handler: http.DefaultServeMux}
	startGroup(ctx, g, "pprof", pprofAddress, pprofServer, time.Second*5)
}

func startGroup(ctx context.Context, g *errgroup.Group, name, port string, server *http.Server, interruptTimeout time.Duration) {
	g.Go(func() error {
		slog.Info("starting server", "name", name, "port", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		slog.Info("server shut down gracefully", "name", name)
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		slog.Info("shutting down server", "name", name, "port", port)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), interruptTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown server", "name", name, slog.String("error", err.Error()))
			return err
		}
		return nil
	})
}
