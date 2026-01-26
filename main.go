package main

import (
	"github.com/Shivamrai15/activity-safari/routes"
	"github.com/Shivamrai15/activity-safari/utils"
	"github.com/gin-gonic/gin"
)

func main() {

	utils.InitConfig()

	port := utils.GetConfigValue("PORT")
	if err := utils.InitLibSql(); err != nil {
		panic("Failed to initialize database: " + err.Error())
	}

	router := gin.Default()

	// defining middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/api/v2/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "success",
		})
	})

	routes.RegisterSearchRoutes(router)

	router.Run(":" + port)
}
