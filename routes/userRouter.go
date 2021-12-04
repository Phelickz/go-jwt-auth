package routes

import (
	"github.com/Phelickz/go-jwt-auth/controllers"
	// "github.com/Phelickz/go-jwt-auth/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	// r.Use(middleware.Authenticate())
	// r.GET("/users", controllers.GetAllUsers())
	r.GET("/users/:user_id", controllers.GetUser())
}
