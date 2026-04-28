// Package server configura y levanta el servidor HTTP de Neto.
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neto-app/neto/api/internal/config"
	"github.com/neto-app/neto/api/internal/handler"
	"github.com/neto-app/neto/api/internal/middleware"
	"github.com/neto-app/neto/api/internal/repository"
	"github.com/neto-app/neto/api/internal/usecase"
)

// Server encapsula el mux de Chi y la configuración del servidor.
type Server struct {
	mux  *chi.Mux
	cfg  config.Config
	pool *pgxpool.Pool
}

// New crea un nuevo Server con el pool de DB ya inicializado.
func New(cfg config.Config, pool *pgxpool.Pool) *Server {
	s := &Server{
		mux:  chi.NewMux(),
		cfg:  cfg,
		pool: pool,
	}
	s.routes()
	return s
}

// routes registra todos los middlewares globales y las rutas de la API.
func (s *Server) routes() {
	// Middlewares globales
	s.mux.Use(chimiddleware.RequestID)
	s.mux.Use(chimiddleware.RealIP)
	s.mux.Use(chimiddleware.Logger)
	s.mux.Use(chimiddleware.Recoverer)
	s.mux.Use(chimiddleware.Timeout(30 * time.Second))

	// Repositorios
	repos := repository.NewRepositories(s.pool)

	// Use cases
	accountUC := usecase.NewAccountUseCase(repos.Account, repos.Currency)
	transactionUC := usecase.NewTransactionUseCase(repos.Transaction, repos.Account)
	categoryUC := usecase.NewCategoryUseCase(repos.Category)

	// Handlers
	healthH := handler.NewHealthHandler()
	accountH := handler.NewAccountHandler(accountUC)
	transactionH := handler.NewTransactionHandler(transactionUC)
	categoryH := handler.NewCategoryHandler(categoryUC)

	// Ruta pública
	s.mux.Get("/health", healthH.Health)

	// Rutas protegidas por JWT
	s.mux.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(s.cfg.SupabaseJWTSecret))

		r.Route("/api/v1", func(r chi.Router) {
			// Accounts
			r.Post("/accounts", accountH.Create)
			r.Get("/accounts", accountH.List)
			r.Get("/accounts/{id}", accountH.GetByID)

			// Transactions — idempotencia solo en POST
			r.With(middleware.Idempotency(s.pool)).Post("/transactions", transactionH.Create)
			r.Get("/transactions", transactionH.List)

			// Categories
			r.Post("/categories", categoryH.Create)
			r.Get("/categories", categoryH.List)
		})
	})
}

// Run levanta el servidor HTTP y maneja graceful shutdown.
func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Canal para errores del servidor
	errCh := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s (env: %s)", addr, s.cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Esperar señal de cierre o error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-quit:
		log.Println("shutdown signal received")
	case <-ctx.Done():
		log.Println("context cancelled")
	}

	// Graceful shutdown con timeout de 10 segundos
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	log.Println("server stopped")
	return nil
}
