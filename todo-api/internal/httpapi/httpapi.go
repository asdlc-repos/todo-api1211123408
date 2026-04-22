package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/todo-api/todo-api/internal/auth"
	"github.com/todo-api/todo-api/internal/config"
	"github.com/todo-api/todo-api/internal/models"
	"github.com/todo-api/todo-api/internal/service"
)

const sessionCookieName = "session"

type Server struct {
	Auth       *auth.Service
	Todos      *service.TodoService
	Categories *service.CategoryService
	Config     config.Config
}

func NewServer(a *auth.Service, t *service.TodoService, c *service.CategoryService, cfg config.Config) *Server {
	return &Server{Auth: a, Todos: t, Categories: c, Config: cfg}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/v1/health", s.handleHealth)

	mux.HandleFunc("/api/v1/auth/register", s.methodOnly(http.MethodPost, s.handleRegister))
	mux.HandleFunc("/api/v1/auth/login", s.methodOnly(http.MethodPost, s.handleLogin))
	mux.HandleFunc("/api/v1/auth/logout", s.methodOnly(http.MethodPost, s.handleLogout))
	mux.HandleFunc("/api/v1/auth/password-reset/request", s.methodOnly(http.MethodPost, s.handlePasswordResetRequest))
	mux.HandleFunc("/api/v1/auth/password-reset/confirm", s.methodOnly(http.MethodPost, s.handlePasswordResetConfirm))
	mux.HandleFunc("/api/v1/auth/me", s.requireAuth(s.handleMe))

	mux.HandleFunc("/api/v1/todos", s.requireAuth(s.handleTodos))
	mux.HandleFunc("/api/v1/todos/", s.requireAuth(s.handleTodoByID))

	mux.HandleFunc("/api/v1/categories", s.requireAuth(s.handleCategories))
	mux.HandleFunc("/api/v1/categories/", s.requireAuth(s.handleCategoryByID))

	return s.withCORS(s.withLogging(mux))
}

func (s *Server) methodOnly(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		next(w, r)
	}
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		serverLog(r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type ctxKey string

const userIDKey ctxKey = "userID"

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		userID, ok := s.Auth.GetUserIDBySession(cookie.Value)
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		r.Header.Set("X-User-ID", userID)
		next(w, r)
	}
}

func currentUserID(r *http.Request) string {
	return r.Header.Get("X-User-ID")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := s.Auth.Register(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid email or password")
			return
		}
		if errors.Is(err, auth.ErrEmailTaken) {
			writeError(w, http.StatusBadRequest, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]string{"status": "registered"})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sess, err := s.Auth.Login(req.Email, req.Password, clientIP(r))
	if err != nil {
		if errors.Is(err, auth.ErrAccountLocked) {
			writeError(w, http.StatusLocked, "account locked")
			return
		}
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sess.Token,
		Path:     "/",
		Expires:  sess.ExpiresAt,
		MaxAge:   int(time.Until(sess.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   s.Config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = s.Auth.Logout(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.Config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

type resetRequestBody struct {
	Email string `json:"email"`
}

func (s *Server) handlePasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var req resetRequestBody
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, _ = s.Auth.RequestReset(req.Email)
	w.WriteHeader(http.StatusAccepted)
	writeJSON(w, map[string]string{"status": "accepted"})
}

type resetConfirmBody struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (s *Server) handlePasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var req resetConfirmBody
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := s.Auth.ConfirmReset(req.Token, req.Password); err != nil {
		if errors.Is(err, auth.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid password")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid token")
		return
	}
	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]string{"status": "updated"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	u, err := s.Auth.Users.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, map[string]string{"id": u.ID, "email": u.Email})
}

func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	switch r.Method {
	case http.MethodGet:
		filter, err := parseTodoFilter(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		todos, err := s.Todos.List(userID, filter)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, todos)
	case http.MethodPost:
		var in service.TodoInput
		if err := decodeJSON(r, &in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		todo, err := s.Todos.Create(userID, in)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, todo)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleTodoByID(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/todos/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		t, err := s.Todos.Get(userID, id)
		if err != nil {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSON(w, t)
	case http.MethodPut:
		var in service.TodoInput
		if err := decodeJSON(r, &in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		t, err := s.Todos.Update(userID, id, in)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, t)
	case http.MethodDelete:
		if err := s.Todos.Delete(userID, id); err != nil {
			writeServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type categoryBody struct {
	Name string `json:"name"`
}

func (s *Server) handleCategories(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	switch r.Method {
	case http.MethodGet:
		cs, err := s.Categories.List(userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, cs)
	case http.MethodPost:
		var body categoryBody
		if err := decodeJSON(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		c, err := s.Categories.Create(userID, body.Name)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, c)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleCategoryByID(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/categories/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var body categoryBody
		if err := decodeJSON(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		c, err := s.Categories.Update(userID, id, body.Name)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, c)
	case http.MethodDelete:
		if err := s.Categories.Delete(userID, id); err != nil {
			writeServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func parseTodoFilter(r *http.Request) (models.TodoFilter, error) {
	q := r.URL.Query()
	var f models.TodoFilter
	if v := q.Get("completed"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return f, errors.New("invalid completed")
		}
		f.Completed = &b
	}
	if v := q.Get("categoryId"); v != "" {
		cid := v
		f.CategoryID = &cid
	}
	if v := q.Get("dueFrom"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, errors.New("invalid dueFrom")
		}
		f.DueFrom = &t
	}
	if v := q.Get("dueTo"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, errors.New("invalid dueTo")
		}
		f.DueTo = &t
	}
	return f, nil
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	return r.RemoteAddr
}

func decodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, "validation error")
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrVersionConflict):
		writeError(w, http.StatusConflict, "version conflict")
	case errors.Is(err, service.ErrDuplicate):
		writeError(w, http.StatusConflict, "duplicate")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func serverLog(method, path string, status int, dur time.Duration) {
	if !strings.HasPrefix(path, "/api/") && path != "/health" {
		return
	}
	_ = method
	_ = status
	_ = dur
}
