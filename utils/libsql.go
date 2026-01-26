package utils

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var (
	dbInstance *sql.DB
	once       sync.Once
)

func InitLibSql() error {
	var initErr error

	once.Do(func() {
		dbURL := GetConfigValue("TURSO_DB_URL")
		authToken := GetConfigValue("TURSO_TOKEN")

		if dbURL == "" || authToken == "" {
			initErr = fmt.Errorf("TURSO_DB_URL and TURSO_TOKEN must be set")
			return
		}

		connStr := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

		db, err := sql.Open("libsql", connStr)
		if err != nil {
			initErr = fmt.Errorf("failed to open database: %w", err)
			return
		}

		if err := db.Ping(); err != nil {
			initErr = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		dbInstance = db
		log.Println("Database: Connected to Turso successfully")
	})

	return initErr
}

func GetDB() *sql.DB {
	if dbInstance == nil {
		log.Fatal("Database not initialized. Call Init() first.")
	}
	return dbInstance
}

func Close() {
	if dbInstance != nil {
		dbInstance.Close()
		log.Println("Database: Connection closed")
	}
}
