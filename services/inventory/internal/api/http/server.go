package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server — HTTP-сервер инвентори.
type Server struct {
	srv *http.Server
}

// NewServer собирает роутер и регистрирует все хендлеры.
func NewServer(
	addr string,
	hosts *HostsHandler,
	projects *ProjectsHandler,
	namespaces *NamespacesHandler,
) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Health-check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Hosts
		r.Route("/hosts", func(r chi.Router) {
			r.Get("/", hosts.ListHosts)
			r.Post("/", hosts.RegisterHost)
			r.Get("/{hostID}", hosts.GetHost)
		})

		// Projects
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projects.ListProjects)
			r.Post("/", projects.CreateProject)
			r.Get("/{projectID}", projects.GetProject)

			// Namespaces вложены в проект
			r.Route("/{projectID}/namespaces", func(r chi.Router) {
				r.Get("/", namespaces.ListNamespaces)
				r.Post("/", namespaces.CreateNamespace)
			})
		})
	})

	return &Server{
		srv: &http.Server{
			Addr:         addr,
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Run запускает HTTP-сервер (блокирующий вызов).
func (s *Server) Run() error {
	log.Printf("http server listening on %s", s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown выполняет graceful shutdown с таймаутом.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
