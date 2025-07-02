package middleware

import (
	"go-multiplayer-quiz-project/backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddeleware(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")
	if token == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization failed, no token found"})
		return
	}

	err := utils.ValidateToken(token, context)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization failed, invalid token"})
		return
	}

}
