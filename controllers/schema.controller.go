package controllers

import (
	"fmt"

	"github.com/Shivamrai15/activity-safari/utils"
)

func SeedSchema() {
	db := utils.GetDb()

	fmt.Println("Seeding database schema...")

	_, err := db.Exec(`
		ALTER TABLE SearchHistory ADD COLUMN user_id TEXT;
		CREATE INDEX idx_search_history_user_id ON SearchHistory(user_id);
	`)
	// _, err := db.Exec(`
	// 	CREATE TABLE IF NOT EXISTS SearchHistory (
	// 		id TEXT PRIMARY KEY,
	// 		name TEXT NOT NULL,
	// 		image TEXT NOT NULL,
	// 		content_id TEXT NOT NULL,
	// 		type TEXT NOT NULL CHECK (type IN ('ARTIST', 'ALBUM', 'SONG', 'PLAYLIST')),
	// 		created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
	// 	);
	// `)
	if err != nil {
		panic("Failed to seed schema: " + err.Error())
	}

	fmt.Println("Database schema seeded successfully")
}