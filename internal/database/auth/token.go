package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"os"
	"time"
)

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func GenerateToken(username string, email string) (string, error) {
	// Create a new JWT token with claims
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
	
	// Print information about the created token
	log.Printf("Token claims added: %+v\n", claims)
	return tokenString, nil
}

func ValidateToken(signedToken string) error {
	// Parse the token
	token, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	
	if err != nil {
		return err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return errors.New("invalid token or claims")
	}
	
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return errors.New("token has expired")
		}
	}
	
	// Print information about the parsed token
	log.Printf("Token claims: %+v\n", token.Claims)
	return nil
}
