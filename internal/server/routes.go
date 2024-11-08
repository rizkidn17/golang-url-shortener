package server

import (
	"encoding/json"
	"golang-url-shortener/internal/database/handler"
	customMiddleware "golang-url-shortener/internal/middleware"
	"log"
	"net/http"
	
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	
	r.Route("/user", func(r chi.Router) {
		r.Post("/register", handler.RegisterUserHandler)
		r.Post("/get-token", handler.GenerateUserTokenHandler)
	})
	
	//grouping routes
	r.Route("/api", func(r chi.Router) {
		r.Use(customMiddleware.AuthorizationHandler)
		r.Get("/", s.HelloWorldHandler)
		r.Route("/v1", func(r chi.Router) {
			r.Get("/", s.HelloWorldHandler)
			
			r.Get("/health", s.healthHandler)
			
			r.Route("/shorten", func(r chi.Router) {
				r.Get("/{shortCode}", handler.GetShortenUrlByShortCodeHandler)
				r.Get("/{shortCode}/stats", handler.GetShortenUrlStatsByShortCodeHandler)
				r.Post("/", handler.CreateShortenUrlHandler)
				r.Put("/{shortCode}", handler.UpdateShortenUrlHandler)
				r.Delete("/{shortCode}", handler.DeleteShortenUrlByShortCodeHandler)
			})
		})
	})
	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"
	
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}
	
	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
