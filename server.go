package main

import (
	"lexicon/go-template/module"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
)

type AppHttpServer struct {
	router *chi.Mux
	cfg    config
}

func NewAppHttpServer(cfg config) (*AppHttpServer, error) {

	r := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		// AllowedOrigins: []string{"https://bo.lexicon.id", "http://localhost:3000"},
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-KEY", "X-ACCESS-TIME", "X-REQUEST-SIGNATURE", "X-API-USER", "X-REQUEST-IDENTITY"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// // Set a timeout value on the request context (ctx), that will signal
	// // through ctx.Done() that the request has timed out and further
	// // processing should be stopped.
	r.Use(middleware.Timeout(2 * time.Minute))

	server := &AppHttpServer{
		router: r,
		cfg:    cfg,
	}
	return server, nil

}

func (s *AppHttpServer) setupRoute() {
	r := s.router
	// cfg := s.cfg
	r.Route("/v1", func(r chi.Router) {
		// r.Use(middlewares.AccessTime())
		// r.Use(middlewares.ApiKey(cfg.BackendApiKey, cfg.ServerSalt))
		// r.Use(middlewares.RequestSignature(cfg.ServerSalt))
		r.Mount("/module", module.Router())
	})
}

func (s *AppHttpServer) start() {
	r := s.router
	cfg := s.cfg
	log.Info().Msg("Starting up server...")

	if err := http.ListenAndServe(cfg.Listen.Addr(), r); err != nil {
		log.Fatal().Err(err).Msg("Failed to start the server")
		return
	}

	log.Info().Msg("Server Stopped")
}
