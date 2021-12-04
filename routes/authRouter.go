package routes

import (
	"github.com/Phelickz/go-jwt-auth/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {
	r.POST("/users/signup", controllers.Signup())
	r.POST("users/login", controllers.Login())

}
