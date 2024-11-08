package handler

import (
	"encoding/json"
	"golang-url-shortener/internal/database"
	"golang-url-shortener/internal/database/auth"
	"golang-url-shortener/internal/database/model"
	"log"
	"net/http"
)

func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	
	var user model.Users
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	
	if err := user.HashPassword(user.Password); err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	
	dbService := database.New()
	db := dbService.ToGormDB()
	
	if err := db.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "User created successfully"}); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}

func GenerateUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	
	var user model.Users
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	
	dbService := database.New()
	db := dbService.ToGormDB()
	
	var existingUser model.Users
	if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
		log.Printf("Error querying database: %v", err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}
	
	if err := existingUser.CheckPassword(user.Password); err != nil {
		log.Printf("Error comparing password: %v", err)
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}
	
	token, err := auth.GenerateToken(existingUser.Username, existingUser.Email)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	
	existingUser.Token = token
	if err := db.Save(&existingUser).Error; err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"username": existingUser.Username, "token": token}); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}
