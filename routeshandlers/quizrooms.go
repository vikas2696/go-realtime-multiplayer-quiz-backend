package routeshandlers

import (
	"go-multiplayer-quiz-project/backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func showAllQuizRooms(context *gin.Context) {

	quizRooms, err := models.GetQuizRoomsFromDB()

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Quizrooms not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"quizrooms": quizRooms})

}

func createQuizRoom(context *gin.Context) {

	var quizRoom models.QuizRoom
	err := context.ShouldBindJSON(&quizRoom)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inputs"})
		return
	}

	generatedId, err := quizRoom.SaveQuizRoomToDB()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create quizroom"})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"quiz_id": generatedId, "message": "Quiz room created successfully!"})
}

func joinQuizRoom(context *gin.Context) {

	var player models.Player

	player_id, found := context.Get("userId")
	if !found {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Unable to authenticate"})
		return
	}

	player_username, found := context.Get("username")
	if !found {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Unable to authenticate"})
		return
	}

	player.PlayerId = player_id.(int64)
	player.Username = player_username.(string)

	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code"})
		return
	}

	err = player.AddPlayerToQuiz(quizId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join: " + err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "Joined Successfully"})

}

func getQuizRoom(context *gin.Context) {

	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "Quiz Room not found"})
		return
	}

	var quizRoom models.QuizRoom
	err = quizRoom.GetQuizRoomFromId(quizId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "Quiz Room not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"quizroom": quizRoom})

}

func leaveQuizRoom(context *gin.Context) {

	var player models.Player

	player_id, found := context.Get("userId")
	if !found {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to authenticate"})
		return
	}

	player_username, found := context.Get("username")
	if !found {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to authenticate"})
		return
	}

	player.PlayerId = player_id.(int64)
	player.Username = player_username.(string)

	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = player.DeletePlayerFromQuiz(quizId)
	if err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "player left the quizroom successfully"})

}

func deleteQuizRoom(context *gin.Context) {

	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
		return
	}

	playerId, found := context.Get("userId")
	if !found {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to authenticate"})
		return
	}

	host := models.IsHost(quizId, playerId.(int64))
	if !host {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Only host can delete"})
		return
	}

	err = models.DeleteQuizRoomFromDB(int64(quizId))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Deletion failed: " + err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "QuizRoom deleted"})

}

func updateScoreSheet(context *gin.Context) {

	var newScoreSheet map[int64]int
	err := context.ShouldBindJSON(&newScoreSheet)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Wrong format"})
		return
	}

	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz room"})
		return
	}

	var quizRoom models.QuizRoom
	err = quizRoom.GetQuizRoomFromId(quizId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz room"})
		return
	}

	err = models.UpdateScoreSheetinDB(int64(quizId), newScoreSheet)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update score in DB"})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "Score Sheet Updated"})

}

func getScoreSheet(context *gin.Context) {

	quizId, err := strconv.Atoi(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz room"})
		return
	}

	var quizRoom models.QuizRoom
	err = quizRoom.GetQuizRoomFromId(quizId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz room"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"score_sheet": quizRoom.ScoreSheet})

}
