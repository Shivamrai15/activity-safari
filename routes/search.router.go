package routes

import (
	"github.com/Shivamrai15/rs-safari/controllers"
	"github.com/Shivamrai15/rs-safari/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterSearchRoutes(router *gin.Engine) {
	router.Use(middlewares.AuthMiddleware())
	router.POST("/api/v2/search", controllers.AddSearchEntry())
	router.GET("/api/v2/search/recent", controllers.GetRecentSearches())
	router.DELETE("/api/v2/search/:id", controllers.DeleteSearchEntry())
}
