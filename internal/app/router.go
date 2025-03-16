package app

import (
	"github.com/gin-gonic/gin"
	"github.com/lemavisaitov/lk-api/internal/handler"
	"github.com/lemavisaitov/lk-api/internal/middleware"
)

func GetRouter(handler *handler.Handle) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.HttpStatusMetric())

	router.POST("/user/signup", handler.Signup)
	router.GET("user/:id", handler.GetUser)
	router.POST("/user/login", handler.Login)
	router.PUT("/user/:id", handler.UpdateUser)
	router.DELETE("/user/:id", handler.DeleteUser)

	return router
}
