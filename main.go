package main

import (
	"go-multiplayer-quiz-project/backend/database"
	"go-multiplayer-quiz-project/backend/routeshandlers"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	database.InitDB()
	server := gin.Default()

	//CORS middleware
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://vikas2696.github.io", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routeshandlers.RunRoutes(server)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // for local dev
	}
	server.Run(":" + port)

}
