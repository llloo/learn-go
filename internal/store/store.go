package store

import (
	"context"
	"fmt"
	"sync"
	"taskapi/internal/task"
	"time"
)

var ErrNotFound = fmt.Errorf("task not found")

type TaskStore interface {
	GetAll(ctx context.Context) ([]task.Task, error)
	GetByID(ctx context.Context, id int) (task.Task, error)
	Create(ctx context.Context, title string) (task.Task, error)
	PartialUpdate(ctx context.Context, id int, title *string, completed *bool) (task.Task, error)
	Delete(ctx context.Context, id int) error
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
		return task.Task{}, ErrNotFound
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

func (s *Store) Delete(ctx context.Context, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return ErrNotFound
	}

	delete(s.tasks, id)
	return nil
}

func (s *Store) PartialUpdate(ctx context.Context, id int, title *string, completed *bool) (task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tasks[id]
	if !exists {
		return task.Task{}, ErrNotFound
	}
	if title != nil {
		t.Title = *title
	}
	if completed != nil {
		t.Completed = *completed
	}
	s.tasks[id] = t
	return t, nil
}
