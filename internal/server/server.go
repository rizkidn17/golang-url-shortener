package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	
	_ "github.com/joho/godotenv/autoload"
	
	"golang-url-shortener/internal/database"
)

type Server struct {
	port int
	
	db database.Service
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	
	dbService := database.New()
	if err := dbService.Migrate(); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
	
	NewServer := &Server{
		port: port,
		
		db: dbService,
	}
	
	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	
	return server
}
