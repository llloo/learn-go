package main

import (
	"log"
	"net/http"
	"taskapi/internal/handler"
	"taskapi/internal/store"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Use(handler.Logger)

	ts := store.NewStore()
	server := &handler.Server{Store: ts}

	r.Get("/tasks", server.HandleGetTasks)
	r.Post("/tasks", server.HandleCreateTask)
	r.Get("/tasks/{id}", server.HandleGetTaskByID)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
