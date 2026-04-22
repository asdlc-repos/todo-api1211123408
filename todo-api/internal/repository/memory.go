package repository

import (
	"strings"
	"sync"
	"time"

	"github.com/todo-api/todo-api/internal/models"
)

type memUserRepo struct {
	mu       sync.RWMutex
	byID     map[string]*models.User
	byEmail  map[string]string
}

func NewMemoryUserRepository() UserRepository {
	return &memUserRepo{byID: map[string]*models.User{}, byEmail: map[string]string{}}
}

func (r *memUserRepo) Create(u *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := strings.ToLower(u.Email)
	if _, ok := r.byEmail[key]; ok {
		return ErrDuplicate
	}
	r.byID[u.ID] = u
	r.byEmail[key] = u.ID
	return nil
}

func (r *memUserRepo) GetByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[strings.ToLower(email)]
	if !ok {
		return nil, ErrNotFound
	}
	u, ok := r.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (r *memUserRepo) GetByID(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (r *memUserRepo) Update(u *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[u.ID]; !ok {
		return ErrNotFound
	}
	r.byID[u.ID] = u
	return nil
}

type memSessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]*models.Session
}

func NewMemorySessionRepository() SessionRepository {
	return &memSessionRepo{sessions: map[string]*models.Session{}}
}

func (r *memSessionRepo) Create(s *models.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.Token] = s
	return nil
}

func (r *memSessionRepo) Get(token string) (*models.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[token]
	if !ok {
		return nil, ErrNotFound
	}
	if time.Now().After(s.ExpiresAt) {
		return nil, ErrNotFound
	}
	return s, nil
}

func (r *memSessionRepo) Delete(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, token)
	return nil
}

func (r *memSessionRepo) DeleteExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for k, s := range r.sessions {
		if now.After(s.ExpiresAt) {
			delete(r.sessions, k)
		}
	}
}

type memResetRepo struct {
	mu     sync.RWMutex
	tokens map[string]*models.PasswordResetToken
}

func NewMemoryPasswordResetRepository() PasswordResetRepository {
	return &memResetRepo{tokens: map[string]*models.PasswordResetToken{}}
}

func (r *memResetRepo) Create(t *models.PasswordResetToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[t.Token] = t
	return nil
}

func (r *memResetRepo) Get(token string) (*models.PasswordResetToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tokens[token]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (r *memResetRepo) MarkUsed(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tokens[token]
	if !ok {
		return ErrNotFound
	}
	t.Used = true
	return nil
}

type memTodoRepo struct {
	mu    sync.RWMutex
	todos map[string]*models.Todo
}

func NewMemoryTodoRepository() TodoRepository {
	return &memTodoRepo{todos: map[string]*models.Todo{}}
}

func (r *memTodoRepo) Create(t *models.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.todos[t.ID] = t
	return nil
}

func (r *memTodoRepo) GetByID(userID, id string) (*models.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.todos[id]
	if !ok || t.UserID != userID {
		return nil, ErrNotFound
	}
	return t, nil
}

func (r *memTodoRepo) List(userID string, f models.TodoFilter) ([]*models.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.Todo, 0)
	for _, t := range r.todos {
		if t.UserID != userID {
			continue
		}
		if f.Completed != nil && t.Completed != *f.Completed {
			continue
		}
		if f.CategoryID != nil {
			if t.CategoryID == nil || *t.CategoryID != *f.CategoryID {
				continue
			}
		}
		if f.DueFrom != nil {
			if t.DueDate == nil || t.DueDate.Before(*f.DueFrom) {
				continue
			}
		}
		if f.DueTo != nil {
			if t.DueDate == nil || t.DueDate.After(*f.DueTo) {
				continue
			}
		}
		out = append(out, t)
	}
	return out, nil
}

func (r *memTodoRepo) Update(t *models.Todo, expectedVersion int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.todos[t.ID]
	if !ok || existing.UserID != t.UserID {
		return ErrNotFound
	}
	if existing.Version != expectedVersion {
		return ErrVersionMismatch
	}
	r.todos[t.ID] = t
	return nil
}

func (r *memTodoRepo) Delete(userID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.todos[id]
	if !ok || t.UserID != userID {
		return ErrNotFound
	}
	delete(r.todos, id)
	return nil
}

func (r *memTodoRepo) UnassignCategory(userID, categoryID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	for _, t := range r.todos {
		if t.UserID != userID {
			continue
		}
		if t.CategoryID != nil && *t.CategoryID == categoryID {
			t.CategoryID = nil
			t.Version++
			t.UpdatedAt = now
		}
	}
}

type memCategoryRepo struct {
	mu         sync.RWMutex
	categories map[string]*models.Category
}

func NewMemoryCategoryRepository() CategoryRepository {
	return &memCategoryRepo{categories: map[string]*models.Category{}}
}

func (r *memCategoryRepo) Create(c *models.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	lower := strings.ToLower(c.Name)
	for _, existing := range r.categories {
		if existing.UserID == c.UserID && strings.ToLower(existing.Name) == lower {
			return ErrDuplicate
		}
	}
	r.categories[c.ID] = c
	return nil
}

func (r *memCategoryRepo) GetByID(userID, id string) (*models.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.categories[id]
	if !ok || c.UserID != userID {
		return nil, ErrNotFound
	}
	return c, nil
}

func (r *memCategoryRepo) GetByName(userID, name string) (*models.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lower := strings.ToLower(name)
	for _, c := range r.categories {
		if c.UserID == userID && strings.ToLower(c.Name) == lower {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (r *memCategoryRepo) List(userID string) ([]*models.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*models.Category, 0)
	for _, c := range r.categories {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *memCategoryRepo) Update(c *models.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.categories[c.ID]
	if !ok || existing.UserID != c.UserID {
		return ErrNotFound
	}
	lower := strings.ToLower(c.Name)
	for _, other := range r.categories {
		if other.ID == c.ID {
			continue
		}
		if other.UserID == c.UserID && strings.ToLower(other.Name) == lower {
			return ErrDuplicate
		}
	}
	r.categories[c.ID] = c
	return nil
}

func (r *memCategoryRepo) Delete(userID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.categories[id]
	if !ok || c.UserID != userID {
		return ErrNotFound
	}
	delete(r.categories, id)
	return nil
}
