package routeshandlers

import (
	"go-multiplayer-quiz-project/backend/middleware"

	"github.com/gin-gonic/gin"
)

func RunRoutes(server *gin.Engine) {

	server.POST("/signup", signUp)
	server.POST("/login", logIn)
	server.GET("/quizrooms", showAllQuizRooms)

	server.GET("/quizrooms/:id/ws/lobby", webSocketLobby)
	server.GET("/quizrooms/:id/ws/live", webSocketLive)
	server.GET("/:topic/update-questions", handleUpdateQuestions)

	AuthOnlyRoutes := server.Group("/", middleware.AuthMiddeleware)
	{
		AuthOnlyRoutes.POST("/create-quizroom", createQuizRoom)
		AuthOnlyRoutes.PATCH("/quizrooms/:id/join", joinQuizRoom)
		AuthOnlyRoutes.GET("/quizrooms/:id/check-page-access", middleware.CheckPageAccess)

		QuizRoomAuthRoutes := AuthOnlyRoutes.Group("/quizrooms/:id", middleware.QuizJoinMiddleware)
		{
			QuizRoomAuthRoutes.PATCH("/leave", leaveQuizRoom)
			QuizRoomAuthRoutes.DELETE("/delete", deleteQuizRoom)
			QuizRoomAuthRoutes.GET("/lobby", getQuizRoom)

			//QuizRoomAuthRoutes.GET("/ws/lobby", webSocketLobby)

			QuizRoomAuthRoutes.GET("/get-questions", getAllQuestions)
			QuizRoomAuthRoutes.PATCH("/update-scoresheet", updateScoreSheet)
			QuizRoomAuthRoutes.GET("/get-scoresheet", getScoreSheet)

			//not using for now
			QuizRoomAuthRoutes.GET("/:ques_id", loadQuestion)
			QuizRoomAuthRoutes.POST("/save-answer", enterAnswer)
			QuizRoomAuthRoutes.GET("/:ques_id/answer", showAnswer)
		}

	}

}
