package store

import (
	"sync"
	"taskapi/internal/task"
	"time"
)

type TaskStore interface {
	GetAll() []task.Task
	GetByID(id int) (task.Task, bool)
	Create(title string) task.Task
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

func (s *Store) GetAll() []task.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]task.Task, 0, len(s.tasks))
	for id := s.nextID - 1; id > 0; id-- {
		if t, exists := s.tasks[id]; exists {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) GetByID(id int) (task.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tasks[id]
	return t, exists
}

func (s *Store) Create(title string) task.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := task.Task{
		Title:     title,
		ID:        s.nextID,
		CreatedAt: time.Now(),
	}
	s.tasks[t.ID] = t
	s.nextID++
	return t
}
