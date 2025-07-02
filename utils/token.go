package utils

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var secretKey string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	secretKey = os.Getenv("SECRET_KEY")
}

func GetSessionToken(userId int64, username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userId,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 12).Unix(),
	})

	return token.SignedString([]byte(secretKey))
}

func ValidateToken(token string, context *gin.Context) error {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return "", errors.New("invalid token: unauthorized access")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return errors.New("invalid token: unauthorized access " + err.Error())
	}

	if !parsedToken.Valid {
		return errors.New("invalid token: unauthorized access")
	}

	claims := parsedToken.Claims.(jwt.MapClaims)

	context.Set("userId", int64(claims["user_id"].(float64)))
	context.Set("username", claims["username"])

	return nil
}
