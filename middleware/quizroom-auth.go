package middleware

import (
	"go-multiplayer-quiz-project/backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func QuizJoinMiddleware(context *gin.Context) {

	var player models.Player

	player_id, found := context.Get("userId")
	if !found {
		context.JSON(http.StatusForbidden, gin.H{"message": "Invalid Token"})
		context.Abort()
		return
	}

	player_username, found := context.Get("username")
	if !found {
		context.JSON(http.StatusForbidden, gin.H{"message": "Invalid Token"})
		context.Abort()
		return
	}

	player.PlayerId = player_id.(int64)
	player.Username = player_username.(string)

	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusForbidden, gin.H{"message": "Invalid Quiz Room"})
		context.Abort()
		return
	}

	if !joinedQuiz(player, quizId) {
		context.JSON(http.StatusForbidden, gin.H{"error": "You have not joined this quiz"})
		context.Abort()
		return
	}
}

func joinedQuiz(player models.Player, quizId int) bool {

	players, err := models.GetJoinedPlayersList(quizId)
	if err != nil {
		return false
	}

	for index := range players {

		if players[index].PlayerId == player.PlayerId {
			return true
		}

	}

	return false
}
