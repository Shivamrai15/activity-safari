package main

import (
	"github.com/Shivamrai15/activity-safari/routes"
	"github.com/Shivamrai15/activity-safari/utils"
	"github.com/gin-gonic/gin"
)

func main() {

	utils.InitConfig()
	utils.InitDb()

	// setting up the router

	port := utils.GetConfigValue("PORT")
	router := gin.Default()

	// defining middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/api/v2/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status":  "healthy",
			"service": "User Activity Service",
			"version": "1.0.0",
		})
	})

	routes.RegisterSearchRoutes(router)

	router.Run(":" + port)
}
