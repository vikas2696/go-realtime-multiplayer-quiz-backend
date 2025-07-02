package middleware

import (
	"go-multiplayer-quiz-project/backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CheckPageAccess(context *gin.Context) {

	var player models.Player

	player_id, found := context.Get("userId")
	if !found {
		context.JSON(http.StatusForbidden, gin.H{"message": "Invalid Token", "allowed": false, "redirectTo": "/quizrooms"})
		context.Abort()
		return
	}

	player_username, found := context.Get("username")
	if !found {
		context.JSON(http.StatusForbidden, gin.H{"message": "Invalid Token", "allowed": false, "redirectTo": "/quizrooms"})
		context.Abort()
		return
	}

	player.PlayerId = player_id.(int64)
	player.Username = player_username.(string)

	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusForbidden, gin.H{"message": "Invalid Quiz Room", "allowed": false, "redirectTo": "/quizrooms"})
		context.Abort()
		return
	}

	if !joinedQuiz(player, quizId) {
		context.JSON(http.StatusForbidden, gin.H{"error": "You have not joined this quiz", "allowed": false, "redirectTo": "/quizrooms"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Access granted", "allowed": true})
}
