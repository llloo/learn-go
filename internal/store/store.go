package store

import (
	"context"
	"fmt"
	"sync"
	"taskapi/internal/task"
	"time"
)

type TaskStore interface {
	GetAll(ctx context.Context) ([]task.Task, error)
	GetByID(ctx context.Context, id int) (task.Task, error)
	Create(ctx context.Context, title string) (task.Task, error)
}

type Store struct {
	mu     sync.Mutex
	tasks  map[int]task.Task
	nextID int
}

func NewStore() *Store {
	return &Store{
		tasks:  make(map[int]task.Task),
		nextID: 1,
	}
}

func (s *Store) GetAll(ctx context.Context) ([]task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]task.Task, 0, len(s.tasks))
	for id := s.nextID - 1; id > 0; id-- {
		if t, exists := s.tasks[id]; exists {
			result = append(result, t)
		}
	}
	return result, nil
}

func (s *Store) GetByID(ctx context.Context, id int) (task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tasks[id]
	if !exists {
		return task.Task{}, fmt.Errorf("task not found")
	}
	return t, nil
}

func (s *Store) Create(ctx context.Context, title string) (task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := task.Task{
		Title:     title,
		ID:        s.nextID,
		CreatedAt: time.Now(),
	}
	s.tasks[t.ID] = t
	s.nextID++
	return t, nil
}
