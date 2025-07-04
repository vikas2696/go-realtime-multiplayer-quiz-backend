package routeshandlers

import (
	"go-multiplayer-quiz-project/backend/models"
	"go-multiplayer-quiz-project/backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func signUp(context *gin.Context) {

	var user models.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user.Password, err = utils.HashPassword(user.Password)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	err = user.SaveUserToDB()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"user_id": user.UserId})

}

func logIn(context *gin.Context) {

	var user models.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err = user.ValidateLogin()
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GetSessionToken(user.UserId, user.Username)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log in"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"token": token})

}
