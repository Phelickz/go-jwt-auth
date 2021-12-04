package main

import (
	"log"
	"os"

	"github.com/Phelickz/go-jwt-auth/database"
	"github.com/Phelickz/go-jwt-auth/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// getting environment

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error Loading Env File")
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	//setting router
	router := gin.New()
	router.Use(gin.Logger())

	//connecting to database
	database.DBinstance()

	//initializing routes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Success"})
	})

	//starting server
	router.Run(":" + port)
}
