package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"golang-url-shortener/internal/database"
	"golang-url-shortener/internal/database/auth"
	"golang-url-shortener/internal/database/model"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func GetShortenUrlByShortCodeHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	
	dbService := database.New()
	db := dbService.ToGormDB()
	
	var shorten model.Shortens
	
	if err := db.Model(&model.Shortens{}).
		First(&shorten).
		Where("short_code = ?", shortCode).
		Select("id", "url", "short_code", "created_at", "updated_at").
		Error; err != nil {
		log.Printf("Error querying database: %v", err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(shorten); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}

func GetShortenUrlStatsByShortCodeHandler(w http.ResponseWriter, r *http.Request) {

}

func CreateShortenUrlHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	
	var shorten model.Shortens
	
	if err := decoder.Decode(&shorten); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := shorten.GenerateShortCode(); err != nil {
		log.Printf("Error generating short code: %v", err)
		http.Error(w, "Failed to generate short code", http.StatusInternalServerError)
		return
	}
	
	// Extract user ID from the token
	token := r.Header.Get("Authorization") // Assumes token is sent in the Authorization header
	if token == "" {
		http.Error(w, "Authorization token is missing", http.StatusUnauthorized)
		return
	}
	
	userId, err := auth.GetUserIdFromToken(token)
	if err != nil {
		log.Printf("Error extracting user from token: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	
	shorten.UserId = userId
	
	dbService := database.New()
	db := dbService.ToGormDB()
	
	if err := db.Create(&shorten).Error; err != nil {
		log.Printf("Error creating shorten: %v", err)
		http.Error(w, "Failed to create shorten", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "Shorten Url created successfully", "full url": shorten.Url, "shorten url": shorten.ShortCode}); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}

func UpdateShortenUrlHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	shortCode := chi.URLParam(r, "shortCode")
	
	var updatedFields map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updatedFields); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	dbService := database.New()
	db := dbService.ToGormDB()
	
	var shorten model.Shortens
	if err := db.Where("short_code = ?", shortCode).First(&shorten).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Shorten not found", http.StatusNotFound)
		} else {
			log.Printf("Error finding shorten: %v", err)
			http.Error(w, "Failed to retrieve shorten", http.StatusInternalServerError)
		}
		return
	}
	
	if err := db.Model(&shorten).Where("short_code = ?", shortCode).Updates(updatedFields).Error; err != nil {
		log.Printf("Error updating shorten: %v", err)
		http.Error(w, "Failed to update shorten", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "Shorten updated successfully"}); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}

func DeleteShortenUrlByShortCodeHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	
	dbService := database.New()
	db := dbService.ToGormDB()
	
	var shorten model.Shortens
	
	if err := db.Where("short_code = ?", shortCode).First(&shorten).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Shorten not found", http.StatusNotFound)
		} else {
			log.Printf("Error finding shorten: %v", err)
			http.Error(w, "Failed to retrieve shorten", http.StatusInternalServerError)
		}
	}
	
	if err := db.Where("short_code = ?", shortCode).Delete(&model.Shortens{}).Error; err != nil {
		log.Printf("Error deleting shorten: %v", err)
		http.Error(w, "Failed to delete shorten", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "Shorten deleted successfully"}); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}
