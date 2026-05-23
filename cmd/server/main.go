package main

import (
	"log"
	"net/http"
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
		log.Fatal("Failed to create migrate instance:", err)
	}
	
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No database changes needed")
			return
		}
		log.Fatal("Failed to run migrations:", err)
	}
}

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	migration(cfg.DatabaseURL)

	r := chi.NewRouter()
	r.Use(handler.Logger)

	ts, err := store.NewPostgresStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to create Postgres store:", err)
	}
	server := &handler.Server{Store: ts}

	r.Get("/tasks", server.HandleGetTasks)
	r.Post("/tasks", server.HandleCreateTask)
	r.Get("/tasks/{id}", server.HandleGetTaskByID)

	log.Println("Listening on :" + cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, r))
}
