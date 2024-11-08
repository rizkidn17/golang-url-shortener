package middleware

import (
	"fmt"
	"golang-url-shortener/internal/database/auth"
	"log"
	"net/http"
)

func AuthorizationHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Missing authorization header")
			return
		}
		
		err := auth.ValidateToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println("Invalid token :", err)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
