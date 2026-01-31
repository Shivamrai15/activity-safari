package routes

import (
	"github.com/Shivamrai15/activity-safari/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterTokenRoutes(router *gin.Engine) {
	router.POST("/api/v3/token/rotate", controllers.RotateTokenController())
	router.POST("/api/v3/token/revoke", controllers.RevokeTokenController())
}
