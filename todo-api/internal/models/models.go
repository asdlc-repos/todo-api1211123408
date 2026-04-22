package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`

	FailedAttempts int        `json:"-"`
	LockedUntil    *time.Time `json:"-"`
}

type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

type PasswordResetToken struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
	Used      bool
}

type Todo struct {
	ID          string     `json:"id"`
	UserID      string     `json:"-"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	CategoryID  *string    `json:"categoryId"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completedAt"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Category struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type TodoFilter struct {
	Completed  *bool
	CategoryID *string
	DueFrom    *time.Time
	DueTo      *time.Time
}
