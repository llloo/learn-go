package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"taskapi/internal/config"
	"taskapi/internal/handler"
	"taskapi/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func migration(databaseURL string) {
	// 这里可以放一些数据库迁移的代码，比如使用 golang-migrate 来管理数据库 schema
	// 也可以在这里检查数据库连接是否正常，或者执行一些初始化操作

	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		slog.Error("Failed to create migrate instance", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No database changes needed")
			return
		}
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
}

func main() {
	logHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logHandler)

	slog.SetDefault(logger)

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	migration(cfg.DatabaseURL)

	r := chi.NewRouter()
	r.Use(handler.Logger)

	ts, err := store.NewPostgresStore(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to create Postgres store", "error", err)
		os.Exit(1)
	}
	server := &handler.Server{Store: ts}

	r.Get("/tasks", server.HandleGetTasks)
	r.Post("/tasks", server.HandleCreateTask)
	r.Get("/tasks/{id}", server.HandleGetTaskByID)
	r.Post("/tasks/batch", server.HandleBatchCreateTasks)

	// 优雅退出
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		slog.Info("Starting server on " + cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down server")
	if err := srv.Shutdown(context.Background()); err != nil {
		slog.Error("Failed to shutdown server", "error", err)
		os.Exit(1)
	}
	slog.Info("Server gracefully stopped")
}
