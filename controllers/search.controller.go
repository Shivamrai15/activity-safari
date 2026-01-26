package controllers

import (
	"log"
	"net/http"

	"github.com/Shivamrai15/activity-safari/models"
	"github.com/Shivamrai15/activity-safari/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate = validator.New()

func getUserFromContext(ctx *gin.Context) (*models.TokenPayload, bool) {
	user, exists := ctx.Get("user")
	if !exists || user == nil {
		return nil, false
	}

	tokenPayload, ok := user.(*models.TokenPayload)
	if !ok {
		return nil, false
	}

	return tokenPayload, true
}

func GetRecentSearches() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, ok := getUserFromContext(ctx)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
			})
			return
		}

		rows, err := utils.GetDB().Query(
			`SELECT id, name, image, content_id, type, created_at 
			 FROM SearchHistory 
			 WHERE user_id = ?
			 ORDER BY created_at DESC 
			 LIMIT 10`,
			token.UserID,
		)
		if err != nil {
			log.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			return
		}
		defer rows.Close()

		var searches []models.Search
		for rows.Next() {
			var search models.Search
			err := rows.Scan(
				&search.Id,
				&search.Name,
				&search.Image,
				&search.ContentId,
				&search.Type,
				&search.CreatedAt,
			)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to parse search results",
				})
				return
			}
			searches = append(searches, search)
		}

		if err = rows.Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    searches,
			"message": "Recent searches retrieved successfully",
		})
	}
}

func DeleteSearchEntry() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, ok := getUserFromContext(ctx)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
			})
			return
		}

		id := ctx.Param("id")
		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "ID is required",
			})
			return
		}

		result, err := utils.GetDB().Exec(
			`DELETE FROM SearchHistory WHERE id = ? AND user_id = ?`,
			id, token.UserID,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to delete search entry",
			})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			return
		}

		if rowsAffected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Search entry not found",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Search entry deleted successfully",
		})
	}
}

type AddSearchRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	Image     string `json:"image" validate:"required,url"`
	ContentId string `json:"content_id" validate:"required"`
	Type      string `json:"type" validate:"required,oneof=ARTIST ALBUM SONG PLAYLIST"`
}

func AddSearchEntry() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, ok := getUserFromContext(ctx)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
			})
			return
		}

		var req AddSearchRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		if err := validate.Struct(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		id := uuid.New().String()

		_, err := utils.GetDB().Exec(
			`INSERT INTO SearchHistory (id, name, image, content_id, type, user_id) 
			 VALUES (?, ?, ?, ?, ?, ?)`,
			id, req.Name, req.Image, req.ContentId, req.Type, token.UserID,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to add search entry",
			})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Search entry added successfully",
			"data": gin.H{
				"id": id,
			},
		})
	}
}
