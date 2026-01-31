package utils

import (
	"database/sql"
	"log"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	dbInstance *sql.DB
	once       sync.Once
)

func InitDb() {
	once.Do(func() {
		var err error

		dbInstance, err = sql.Open(
			"sqlite",
			"file:data.db?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)",
		)
		if err != nil {
			log.Fatal("Failed to open database:", err)
		}

		dbInstance.SetMaxOpenConns(1)
		dbInstance.SetMaxIdleConns(1)

		if err = dbInstance.Ping(); err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		pragmas := []string{
			"PRAGMA cache_size = -20000;",
			"PRAGMA temp_store = MEMORY;",
			"PRAGMA busy_timeout = 5000;",
		}

		for _, p := range pragmas {
			if _, err := dbInstance.Exec(p); err != nil {
				log.Fatal("Failed to set pragma:", err)
			}
		}

		log.Println("SQLite database initialized successfully")
	})
}

func GetDb() *sql.DB {
	InitDb()
	return dbInstance
}
