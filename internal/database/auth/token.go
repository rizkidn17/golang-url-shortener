package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang-url-shortener/internal/database"
	"golang-url-shortener/internal/database/model"
	"log"
	"os"
	"time"
)

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func GenerateToken(username string, email string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usr": username,                           // Subject (user identifier)
		"eml": email,                              // User email
		"iss": "golang-url-shortener",             // Issuer
		"exp": time.Now().AddDate(1, 0, 0).Unix(), // Expiration time
		"iat": time.Now().Unix(),                  // Issued at
	})
	
	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	
	log.Printf("Token claims added: %+v\n", claims)
	return tokenString, nil
}

func ParseAndValidateToken(signedToken string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	
	if err != nil {
		return nil, nil, err
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, nil, errors.New("invalid token or claims")
	}
	
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return nil, nil, errors.New("token has expired")
		}
	}
	
	return token, claims, nil
}

func ValidateToken(signedToken string) error {
	_, _, err := ParseAndValidateToken(signedToken)
	if err != nil {
		return err
	}
	
	log.Printf("Token validated successfully.")
	return nil
}

func GetUserIdFromToken(signedToken string) (uint, error) {
	_, claims, err := ParseAndValidateToken(signedToken)
	if err != nil {
		return 0, err
	}
	
	usr, ok := claims["usr"].(string)
	if !ok {
		return 0, errors.New("username not found in token")
	}
	
	var user model.Users
	dbService := database.New()
	db := dbService.ToGormDB()
	
	if err := db.Where("username = ?", usr).First(&user).Error; err != nil {
		return 0, err
	}
	
	return user.ID, nil
}
