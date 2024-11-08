package model

import (
	"crypto/rand"
	"math/big"
	"os"
	"strconv"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var shortCodeLength = func() int {
	if lengthStr := os.Getenv("SHORT_CODE_LENGTH"); lengthStr != "" {
		if length, err := strconv.Atoi(lengthStr); err == nil {
			return length
		}
	}
	return 6 // default length
}()

type Shortens struct {
	ID        uint   `json:"id" gorm:"auto_increment;unique"`
	Url       string `json:"url"`
	ShortCode string `json:"short_code" gorm:"unique"`
	UserId    uint   `json:"user_id"`
	User      Users  `json:"user" gorm:"foreignKey:UserId;references:ID"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// generateRandomString generates a random string of specified length.
func (shorten *Shortens) GenerateShortCode() error {
	result := make([]byte, shortCodeLength)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return err
		}
		result[i] = charset[num.Int64()]
	}
	shorten.ShortCode = string(result)
	return nil
}
