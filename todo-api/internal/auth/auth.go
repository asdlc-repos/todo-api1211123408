package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/todo-api/todo-api/internal/models"
	"github.com/todo-api/todo-api/internal/repository"
	"github.com/todo-api/todo-api/internal/validation"
)

const (
	MaxFailedAttempts = 3
	LockoutDuration   = 15 * time.Minute
	SessionTTL        = 24 * time.Hour
	ResetTokenTTL     = 1 * time.Hour
)

var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked  = errors.New("account locked")
	ErrInvalidToken   = errors.New("invalid token")
	ErrEmailTaken     = errors.New("email taken")
)

type Service struct {
	Users    repository.UserRepository
	Sessions repository.SessionRepository
	Resets   repository.PasswordResetRepository
}

func NewService(u repository.UserRepository, s repository.SessionRepository, r repository.PasswordResetRepository) *Service {
	return &Service{Users: u, Sessions: s, Resets: r}
}

func (s *Service) Register(email, password string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !validation.ValidEmail(email) {
		return nil, ErrInvalidInput
	}
	if !validation.ValidPassword(password) {
		return nil, ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &models.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.Users.Create(u); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(email, password, ip string) (*models.Session, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.Users.GetByEmail(email)
	if err != nil {
		logAuth("login_fail_no_user", email, ip)
		return nil, ErrInvalidCredentials
	}
	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		logAuth("login_fail_locked", email, ip)
		return nil, ErrAccountLocked
	}
	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		u.FailedAttempts++
		if u.FailedAttempts >= MaxFailedAttempts {
			until := time.Now().Add(LockoutDuration)
			u.LockedUntil = &until
			u.FailedAttempts = 0
			_ = s.Users.Update(u)
			logAuth("login_fail_locked_now", email, ip)
			return nil, ErrAccountLocked
		}
		_ = s.Users.Update(u)
		logAuth("login_fail_bad_password", email, ip)
		return nil, ErrInvalidCredentials
	}
	u.FailedAttempts = 0
	u.LockedUntil = nil
	_ = s.Users.Update(u)

	token, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	sess := &models.Session{
		Token:     token,
		UserID:    u.ID,
		ExpiresAt: time.Now().Add(SessionTTL),
	}
	if err := s.Sessions.Create(sess); err != nil {
		return nil, err
	}
	logAuth("login_success", email, ip)
	return sess, nil
}

func (s *Service) Logout(token string) error {
	return s.Sessions.Delete(token)
}

func (s *Service) RequestReset(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.Users.GetByEmail(email)
	if err != nil {
		return "", nil
	}
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	rt := &models.PasswordResetToken{
		Token:     token,
		UserID:    u.ID,
		ExpiresAt: time.Now().Add(ResetTokenTTL),
	}
	if err := s.Resets.Create(rt); err != nil {
		return "", err
	}
	log.Printf("password_reset_token email=%s token=%s", email, token)
	return token, nil
}

func (s *Service) ConfirmReset(token, newPassword string) error {
	rt, err := s.Resets.Get(token)
	if err != nil {
		return ErrInvalidToken
	}
	if rt.Used || time.Now().After(rt.ExpiresAt) {
		return ErrInvalidToken
	}
	if !validation.ValidPassword(newPassword) {
		return ErrInvalidInput
	}
	u, err := s.Users.GetByID(rt.UserID)
	if err != nil {
		return ErrInvalidToken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	u.FailedAttempts = 0
	u.LockedUntil = nil
	if err := s.Users.Update(u); err != nil {
		return err
	}
	_ = s.Resets.MarkUsed(token)
	return nil
}

func (s *Service) GetUserIDBySession(token string) (string, bool) {
	sess, err := s.Sessions.Get(token)
	if err != nil {
		return "", false
	}
	return sess.UserID, true
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func logAuth(event, email, ip string) {
	log.Printf("auth_event=%s ts=%s email=%s ip=%s", event, time.Now().UTC().Format(time.RFC3339), email, ip)
}
