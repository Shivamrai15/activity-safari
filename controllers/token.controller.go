package controllers

import (
	"net/http"

	"github.com/Shivamrai15/activity-safari/utils"
	"github.com/gin-gonic/gin"
)

type RotateTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func RotateTokenController() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RotateTokenRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "refreshToken is required",
			})
			return
		}

		tokenRotator := utils.NewTokenRotator()

		result, err := tokenRotator.Rotate(ctx.Request.Context(), req.RefreshToken)
		if err != nil {
			statusCode := http.StatusUnauthorized
			errorMessage := err.Error()

			if err == utils.ErrTokenReused {
				statusCode = http.StatusForbidden
				errorMessage = "Token has been revoked or reused"
			} else if err == utils.ErrInvalidToken {
				errorMessage = "Invalid or expired refresh token"
			}

			ctx.JSON(statusCode, gin.H{
				"success": false,
				"error":   errorMessage,
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success":      true,
			"accessToken":  result.AccessToken,
			"refreshToken": result.RefreshToken,
		})
	}
}

type RevokeTokenRequest struct {
	AccessToken string `json:"accessToken" binding:"required"`
}

func RevokeTokenController() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RevokeTokenRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "accessToken is required",
			})
			return
		}

		tokenRotator := utils.NewTokenRotator()

		err := tokenRotator.Revoke(ctx.Request.Context(), req.AccessToken)
		if err != nil {
			statusCode := http.StatusUnauthorized
			errorMessage := err.Error()

			if err == utils.ErrInvalidAccessToken {
				errorMessage = "Invalid or expired access token"
			} else if err == utils.ErrNoActiveSessions {
				statusCode = http.StatusNotFound
				errorMessage = "No active sessions found for this user"
			}

			ctx.JSON(statusCode, gin.H{
				"success": false,
				"error":   errorMessage,
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "All sessions have been revoked successfully",
		})
	}
}
