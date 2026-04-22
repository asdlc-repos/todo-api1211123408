package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/todo-api/todo-api/internal/models"
	"github.com/todo-api/todo-api/internal/repository"
	"github.com/todo-api/todo-api/internal/validation"
)

type CategoryService struct {
	Categories repository.CategoryRepository
	Todos      repository.TodoRepository
}

func NewCategoryService(c repository.CategoryRepository, t repository.TodoRepository) *CategoryService {
	return &CategoryService{Categories: c, Todos: t}
}

func (s *CategoryService) Create(userID, name string) (*models.Category, error) {
	n, ok := validation.Sanitize(name)
	if !ok || n == "" || len(n) > 50 {
		return nil, ErrValidation
	}
	c := &models.Category{
		ID:        uuid.NewString(),
		UserID:    userID,
		Name:      n,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.Categories.Create(c); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) List(userID string) ([]*models.Category, error) {
	return s.Categories.List(userID)
}

func (s *CategoryService) Update(userID, id, name string) (*models.Category, error) {
	existing, err := s.Categories.GetByID(userID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	n, ok := validation.Sanitize(name)
	if !ok || n == "" || len(n) > 50 {
		return nil, ErrValidation
	}
	updated := *existing
	updated.Name = n
	if err := s.Categories.Update(&updated); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrDuplicate
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &updated, nil
}

func (s *CategoryService) Delete(userID, id string) error {
	if _, err := s.Categories.GetByID(userID, id); err != nil {
		return ErrNotFound
	}
	s.Todos.UnassignCategory(userID, id)
	if err := s.Categories.Delete(userID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
