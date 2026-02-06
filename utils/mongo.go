package utils

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	mongoClient      *mongo.Client
	mongoDB          *mongo.Database
	mongoOnce        sync.Once
	mongoInitialized bool
	mongoMutex       sync.RWMutex
)


func InitMongo() {
	mongoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		mongoURI := GetConfigValue("MONGO_URI")
		mongoDatabase := GetConfigValue("MONGO_DATABASE")

		clientOpts := options.Client().
			ApplyURI(mongoURI).
			SetMaxPoolSize(50).                          // Maximum number of connections in the pool
			SetMinPoolSize(5).                           // Minimum number of connections in the pool
			SetMaxConnIdleTime(30 * time.Second).        // Close idle connections after 30s
			SetConnectTimeout(10 * time.Second).         // Connection timeout
			SetServerSelectionTimeout(10 * time.Second). // Server selection timeout
			SetRetryWrites(true).                        // Enable retryable writes
			SetRetryReads(true)                          // Enable retryable reads

		var err error
		mongoClient, err = mongo.Connect(ctx, clientOpts)
		if err != nil {
			log.Fatal("Failed to create MongoDB client:", err)
		}

		// Verify connection
		if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
			log.Fatal("Failed to ping MongoDB:", err)
		}

		mongoDB = mongoClient.Database(mongoDatabase)

		mongoMutex.Lock()
		mongoInitialized = true
		mongoMutex.Unlock()

		log.Println("MongoDB initialized successfully - single persistent connection established")
	})
}


func GetMongoDB() *mongo.Database {
	mongoMutex.RLock()
	initialized := mongoInitialized
	mongoMutex.RUnlock()

	if !initialized {
		log.Fatal("MongoDB not initialized. Call InitMongo() first.")
	}
	return mongoDB
}


func GetMongoClient() *mongo.Client {
	mongoMutex.RLock()
	initialized := mongoInitialized
	mongoMutex.RUnlock()

	if !initialized {
		log.Fatal("MongoDB not initialized. Call InitMongo() first.")
	}
	return mongoClient
}


func CloseMongo() error {
	mongoMutex.RLock()
	initialized := mongoInitialized
	mongoMutex.RUnlock()

	if !initialized || mongoClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mongoClient.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
		return err
	}

	log.Println("MongoDB connection closed successfully")
	return nil
}


func PingMongo() error {
	mongoMutex.RLock()
	initialized := mongoInitialized
	mongoMutex.RUnlock()

	if !initialized || mongoClient == nil {
		return mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return mongoClient.Ping(ctx, readpref.Primary())
}
