package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"taskapi/internal/task"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) GetAll(ctx context.Context) ([]task.Task, error) {
	tasks := make([]task.Task, 0)

	rows, err := s.db.QueryContext(ctx, "SELECT id, title, created_at, completed FROM tasks ORDER BY id DESC")
	if err != nil {
		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		var t task.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.CreatedAt, &t.Completed); err != nil {
			log.Printf("failed to scan row: %v", err)
			continue
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return tasks, err
	}

	return tasks, nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id int) (task.Task, error) {
	var t task.Task
	row := s.db.QueryRowContext(ctx, "SELECT id, title, created_at, completed FROM tasks WHERE id = $1", id)

	if err := row.Scan(&t.ID, &t.Title, &t.CreatedAt, &t.Completed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return t, fmt.Errorf("task not found")
		}
		return t, err
	}

	return t, nil
}

func (s *PostgresStore) Create(ctx context.Context, title string) (task.Task, error) {
	var t task.Task
	row := s.db.QueryRowContext(ctx, "INSERT INTO tasks (title) VALUES ($1) RETURNING id, title, created_at, completed", title)

	if err := row.Scan(&t.ID, &t.Title, &t.CreatedAt, &t.Completed); err != nil {
		return t, err
	}
	return t, nil
}
