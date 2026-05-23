package main

import (
	"log"
	"net/http"
	"taskapi/internal/config"
	"taskapi/internal/handler"
	"taskapi/internal/store"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	r := chi.NewRouter()
	r.Use(handler.Logger)

	ts := store.NewStore()
	server := &handler.Server{Store: ts}

	r.Get("/tasks", server.HandleGetTasks)
	r.Post("/tasks", server.HandleCreateTask)
	r.Get("/tasks/{id}", server.HandleGetTaskByID)

	log.Println("Listening on :" + cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, r))
}
