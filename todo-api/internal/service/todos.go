package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/todo-api/todo-api/internal/models"
	"github.com/todo-api/todo-api/internal/repository"
	"github.com/todo-api/todo-api/internal/validation"
)

var (
	ErrValidation      = errors.New("validation error")
	ErrNotFound        = errors.New("not found")
	ErrVersionConflict = errors.New("version conflict")
	ErrDuplicate       = errors.New("duplicate")
)

type TodoInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	DueDate     *string    `json:"dueDate"`
	CategoryID  *string    `json:"categoryId"`
	Completed   *bool      `json:"completed"`
	Version     *int       `json:"version"`
}

type TodoService struct {
	Todos      repository.TodoRepository
	Categories repository.CategoryRepository
}

func NewTodoService(t repository.TodoRepository, c repository.CategoryRepository) *TodoService {
	return &TodoService{Todos: t, Categories: c}
}

func (s *TodoService) Create(userID string, in TodoInput) (*models.Todo, error) {
	if in.Title == nil {
		return nil, ErrValidation
	}
	title, ok := validation.Sanitize(*in.Title)
	if !ok || title == "" || len(title) > 200 {
		return nil, ErrValidation
	}
	description := ""
	if in.Description != nil {
		d, ok := validation.Sanitize(*in.Description)
		if !ok || len(d) > 2000 {
			return nil, ErrValidation
		}
		description = d
	}
	var due *time.Time
	if in.DueDate != nil && strings.TrimSpace(*in.DueDate) != "" {
		t, err := time.Parse(time.RFC3339, *in.DueDate)
		if err != nil {
			return nil, ErrValidation
		}
		due = &t
	}
	var categoryID *string
	if in.CategoryID != nil && *in.CategoryID != "" {
		if _, err := s.Categories.GetByID(userID, *in.CategoryID); err != nil {
			return nil, ErrValidation
		}
		id := *in.CategoryID
		categoryID = &id
	}
	now := time.Now().UTC()
	todo := &models.Todo{
		ID:          uuid.NewString(),
		UserID:      userID,
		Title:       title,
		Description: description,
		DueDate:     due,
		CategoryID:  categoryID,
		Completed:   false,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if in.Completed != nil && *in.Completed {
		todo.Completed = true
		todo.CompletedAt = &now
	}
	if err := s.Todos.Create(todo); err != nil {
		return nil, err
	}
	return todo, nil
}

func (s *TodoService) Get(userID, id string) (*models.Todo, error) {
	t, err := s.Todos.GetByID(userID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return t, nil
}

func (s *TodoService) List(userID string, f models.TodoFilter) ([]*models.Todo, error) {
	return s.Todos.List(userID, f)
}

func (s *TodoService) Update(userID, id string, in TodoInput) (*models.Todo, error) {
	existing, err := s.Todos.GetByID(userID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	if in.Version == nil {
		return nil, ErrValidation
	}
	if *in.Version != existing.Version {
		return nil, ErrVersionConflict
	}
	updated := *existing

	if in.Title != nil {
		title, ok := validation.Sanitize(*in.Title)
		if !ok || title == "" || len(title) > 200 {
			return nil, ErrValidation
		}
		updated.Title = title
	}
	if in.Description != nil {
		d, ok := validation.Sanitize(*in.Description)
		if !ok || len(d) > 2000 {
			return nil, ErrValidation
		}
		updated.Description = d
	}
	if in.DueDate != nil {
		if strings.TrimSpace(*in.DueDate) == "" {
			updated.DueDate = nil
		} else {
			t, err := time.Parse(time.RFC3339, *in.DueDate)
			if err != nil {
				return nil, ErrValidation
			}
			updated.DueDate = &t
		}
	}
	if in.CategoryID != nil {
		if *in.CategoryID == "" {
			updated.CategoryID = nil
		} else {
			if _, err := s.Categories.GetByID(userID, *in.CategoryID); err != nil {
				return nil, ErrValidation
			}
			cid := *in.CategoryID
			updated.CategoryID = &cid
		}
	}
	now := time.Now().UTC()
	if in.Completed != nil {
		if *in.Completed && !existing.Completed {
			updated.Completed = true
			updated.CompletedAt = &now
		} else if !*in.Completed && existing.Completed {
			updated.Completed = false
			updated.CompletedAt = nil
		}
	}
	updated.Version = existing.Version + 1
	updated.UpdatedAt = now

	if err := s.Todos.Update(&updated, existing.Version); err != nil {
		if errors.Is(err, repository.ErrVersionMismatch) {
			return nil, ErrVersionConflict
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &updated, nil
}

func (s *TodoService) Delete(userID, id string) error {
	if err := s.Todos.Delete(userID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
