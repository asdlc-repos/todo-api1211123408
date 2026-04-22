package repository

import (
	"errors"

	"github.com/todo-api/todo-api/internal/models"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrDuplicate     = errors.New("duplicate")
	ErrVersionMismatch = errors.New("version mismatch")
)

type UserRepository interface {
	Create(u *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetByID(id string) (*models.User, error)
	Update(u *models.User) error
}

type SessionRepository interface {
	Create(s *models.Session) error
	Get(token string) (*models.Session, error)
	Delete(token string) error
	DeleteExpired()
}

type PasswordResetRepository interface {
	Create(t *models.PasswordResetToken) error
	Get(token string) (*models.PasswordResetToken, error)
	MarkUsed(token string) error
}

type TodoRepository interface {
	Create(t *models.Todo) error
	GetByID(userID, id string) (*models.Todo, error)
	List(userID string, filter models.TodoFilter) ([]*models.Todo, error)
	Update(t *models.Todo, expectedVersion int) error
	Delete(userID, id string) error
	UnassignCategory(userID, categoryID string)
}

type CategoryRepository interface {
	Create(c *models.Category) error
	GetByID(userID, id string) (*models.Category, error)
	GetByName(userID, name string) (*models.Category, error)
	List(userID string) ([]*models.Category, error)
	Update(c *models.Category) error
	Delete(userID, id string) error
}
