package app

import "github.com/gin-gonic/gin"

func GetRouter(handler Handler) *gin.Engine {
	router := gin.Default()

	router.POST("/user/signup", handler.Signup)
	router.GET("user/:id", handler.GetUser)
	router.POST("/user/login", handler.Login)
	router.PUT("/user/:id", handler.UpdateUser)
	router.DELETE("/user/:id", handler.DeleteUser)

	return router
}
