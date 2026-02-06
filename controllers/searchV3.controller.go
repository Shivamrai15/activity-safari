package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Shivamrai15/activity-safari/models"
	"github.com/Shivamrai15/activity-safari/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const searchHistoryCollection = "search_history"

type SearchDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Image     string             `bson:"image"`
	ContentID string             `bson:"content_id"`
	Type      string             `bson:"type"`
	UserID    primitive.ObjectID `bson:"user_id"`
	CreatedAt time.Time          `bson:"created_at"`
}

func CreateSearchCollectionV3() error {
	collection := utils.GetMongoDB().Collection(searchHistoryCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("user_id_created_at_idx"),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "content_id", Value: 1},
			},
			Options: options.Index().SetName("user_id_content_id_idx").SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Error creating indexes for search_history: %v", err)
		return err
	}

	log.Println("Successfully created indexes for search_history collection")
	return nil
}

func GetRecentSearchesV3() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, ok := getUserFromContext(ctx)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
			})
			return
		}

		userObjectID, err := primitive.ObjectIDFromHex(token.UserID)
		if err != nil {
			log.Printf("Invalid user_id format: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid user ID format",
			})
			return
		}

		collection := utils.GetMongoDB().Collection(searchHistoryCollection)

		mongoCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"user_id": userObjectID}
		findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).
			SetLimit(10)

		cursor, err := collection.Find(mongoCtx, filter, findOptions)
		if err != nil {
			log.Printf("Error querying MongoDB: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
			})
			return
		}
		defer cursor.Close(mongoCtx)

		var searchDocs []SearchDocument
		if err = cursor.All(mongoCtx, &searchDocs); err != nil {
			log.Printf("Error decoding MongoDB results: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to parse search results",
			})
			return
		}

		searches := make([]models.Search, 0, len(searchDocs))
		for i := range searchDocs {
			doc := &searchDocs[i]
			idHex := doc.ID.Hex()
			searches = append(searches, models.Search{
				Id:        &idHex,
				Name:      &doc.Name,
				Image:     &doc.Image,
				ContentId: &doc.ContentID,
				Type:      &doc.Type,
				CreatedAt: doc.CreatedAt,
			})
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    searches,
			"message": "Recent searches retrieved successfully",
		})
	}
}

func DeleteSearchEntryV3() gin.HandlerFunc {
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

		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			log.Printf("Invalid search entry ID format: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid search entry ID format",
			})
			return
		}

		userObjectID, err := primitive.ObjectIDFromHex(token.UserID)
		if err != nil {
			log.Printf("Invalid user_id format: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid user ID format",
			})
			return
		}

		collection := utils.GetMongoDB().Collection(searchHistoryCollection)

		mongoCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{
			"_id":     objectID,
			"user_id": userObjectID,
		}

		result, err := collection.DeleteOne(mongoCtx, filter)
		if err != nil {
			log.Printf("Error deleting from MongoDB: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to delete search entry",
			})
			return
		}

		if result.DeletedCount == 0 {
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

func AddSearchEntryV3() gin.HandlerFunc {
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

		userObjectID, err := primitive.ObjectIDFromHex(token.UserID)
		if err != nil {
			log.Printf("Invalid user_id format: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid user ID format",
			})
			return
		}

		collection := utils.GetMongoDB().Collection(searchHistoryCollection)

		mongoCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{
			"user_id":    userObjectID,
			"content_id": req.ContentId,
		}

		id := primitive.NewObjectID()
		now := time.Now()

		update := bson.M{
			"$set": bson.M{
				"name":       req.Name,
				"image":      req.Image,
				"type":       req.Type,
				"created_at": now,
			},
			"$setOnInsert": bson.M{
				"_id": id,
			},
		}

		opts := options.Update().SetUpsert(true)
		result, err := collection.UpdateOne(mongoCtx, filter, update, opts)
		if err != nil {
			log.Printf("Error upserting into MongoDB: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to add search entry",
			})
			return
		}

		responseId := id.Hex()
		if result.UpsertedCount == 0 {
			var existingDoc SearchDocument
			err := collection.FindOne(mongoCtx, filter).Decode(&existingDoc)
			if err == nil {
				responseId = existingDoc.ID.Hex()
			}
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Search entry added successfully",
			"data": gin.H{
				"id": responseId,
			},
		})
	}
}

func ClearAllSearchesV3() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, ok := getUserFromContext(ctx)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
			})
			return
		}

		userObjectID, err := primitive.ObjectIDFromHex(token.UserID)
		if err != nil {
			log.Printf("Invalid user_id format: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid user ID format",
			})
			return
		}

		collection := utils.GetMongoDB().Collection(searchHistoryCollection)

		mongoCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"user_id": userObjectID}

		result, err := collection.DeleteMany(mongoCtx, filter)
		if err != nil {
			log.Printf("Error clearing searches from MongoDB: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to clear search history",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Search history cleared successfully",
			"data": gin.H{
				"deleted_count": result.DeletedCount,
			},
		})
	}
}
