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

	"github.com/neto-app/neto/api/internal/ai"
	"github.com/neto-app/neto/api/internal/config"
	"github.com/neto-app/neto/api/internal/handler"
	"github.com/neto-app/neto/api/internal/infrastructure/anthropic"
	"github.com/neto-app/neto/api/internal/infrastructure/openrouter"
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
// Retorna error si no puede alcanzar el JWKS endpoint de Supabase.
func New(cfg config.Config, pool *pgxpool.Pool) (*Server, error) {
	s := &Server{
		mux:  chi.NewMux(),
		cfg:  cfg,
		pool: pool,
	}
	if err := s.routes(); err != nil {
		return nil, err
	}
	return s, nil
}

// routes registra todos los middlewares globales y las rutas de la API.
func (s *Server) routes() error {
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

	// AI Agent — seleccionar cliente según LLM_PROVIDER
	var llmClient ai.LLMClient
	switch s.cfg.LLMProvider {
	case "anthropic":
		llmClient = anthropic.NewClient(s.cfg.AnthropicAPIKey).
			WithSystemPrompt(netoSystemPrompt())
	default: // "openrouter"
		llmClient = openrouter.NewClient(s.cfg.OpenRouterKey, s.cfg.LLMModel).
			WithSystemPrompt(netoSystemPrompt())
	}
	toolExecutor := usecase.NewAPIToolExecutor(usecase.UseCases{
		Transaction: transactionUC,
		Account:     accountUC,
		Category:    categoryUC,
	})
	chatUC := usecase.NewChatUseCase(
		llmClient,
		toolExecutor,
		anthropic.ToolCatalog(),
		repos.AIConversation,
		repos.AIMessage,
	)

	// Handlers
	healthH := handler.NewHealthHandler()
	accountH := handler.NewAccountHandler(accountUC)
	transactionH := handler.NewTransactionHandler(transactionUC)
	categoryH := handler.NewCategoryHandler(categoryUC)
	chatH := handler.NewChatHandler(chatUC)

	// Auth middleware via JWKS (soporta HS256 legacy y ECC P-256)
	authMiddleware, err := middleware.JWKSAuthenticator(s.cfg.JWKSUrl())
	if err != nil {
		return fmt.Errorf("init JWKS authenticator: %w", err)
	}

	// Ruta pública
	s.mux.Get("/health", healthH.Health)

	// Rutas protegidas por JWT
	s.mux.Group(func(r chi.Router) {
		r.Use(authMiddleware)

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

			// Chat — agente de IA
			r.With(middleware.Idempotency(s.pool)).Post("/chat", chatH.Chat)
		})
	})

	return nil
}

// netoSystemPrompt retorna el system prompt del agente Neto en español.
func netoSystemPrompt() string {
	return `Eres Neto, un asistente financiero personal en español neutro.
Ayudas a los usuarios a registrar y analizar sus finanzas personales.

Tus capacidades incluyen:
- Registrar ingresos y egresos
- Consultar saldos de cuentas
- Listar y filtrar transacciones
- Ver resúmenes mensuales
- Registrar deudas (a pagar o cobrar)
- Crear cuentas bancarias o billeteras
- Seguir el progreso de metas de ahorro

Reglas importantes:
- Responde siempre en español neutro (sin voseo ni regionalismos).
- Antes de ejecutar operaciones destructivas o de creación, pide confirmación.
- Cuando el usuario mencione un gasto o ingreso sin especificar la cuenta, pregunta cuál cuenta usar si hay más de una.
- Usa los tools disponibles para leer datos reales del usuario antes de dar resúmenes.
- Formatea los montos con separadores de miles y símbolo de moneda.
- Sé conciso y amigable.`
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
