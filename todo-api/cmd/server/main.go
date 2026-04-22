package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/todo-api/todo-api/internal/auth"
	"github.com/todo-api/todo-api/internal/config"
	"github.com/todo-api/todo-api/internal/httpapi"
	"github.com/todo-api/todo-api/internal/repository"
	"github.com/todo-api/todo-api/internal/service"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)

	cfg := config.Load()

	userRepo := repository.NewMemoryUserRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	resetRepo := repository.NewMemoryPasswordResetRepository()
	todoRepo := repository.NewMemoryTodoRepository()
	categoryRepo := repository.NewMemoryCategoryRepository()

	authSvc := auth.NewService(userRepo, sessionRepo, resetRepo)
	todoSvc := service.NewTodoService(todoRepo, categoryRepo)
	categorySvc := service.NewCategoryService(categoryRepo, todoRepo)

	server := httpapi.NewServer(authSvc, todoSvc, categorySvc, cfg)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			sessionRepo.DeleteExpired()
		}
	}()

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("todo-api listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}
