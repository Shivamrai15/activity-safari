package routes

import (
	"github.com/Shivamrai15/activity-safari/controllers"
	"github.com/Shivamrai15/activity-safari/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterSearchRoutes(router *gin.Engine) {
	router.Use(middlewares.AuthMiddleware())

	// V2 routes - SQLite
	router.POST("/api/v2/search", controllers.AddSearchEntry())
	router.GET("/api/v2/search/recent", controllers.GetRecentSearches())
	router.DELETE("/api/v2/search/:id", controllers.DeleteSearchEntry())

	// V3 routes - MongoDB
	router.POST("/api/v3/search", controllers.AddSearchEntryV3())
	router.DELETE("/api/v3/search", controllers.ClearAllSearchesV3())
	router.GET("/api/v3/search/recent", controllers.GetRecentSearchesV3())
	router.DELETE("/api/v3/search/:id", controllers.DeleteSearchEntryV3())
}
