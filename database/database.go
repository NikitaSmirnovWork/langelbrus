package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	createTables()
	log.Println("Database initialized")
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS modules (
			id        TEXT PRIMARY KEY,
			user_id   TEXT NOT NULL,
			name      TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS topics (
			id         TEXT PRIMARY KEY,
			module_id  TEXT NOT NULL,
			user_id    TEXT NOT NULL,
			name       TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS words (
			id         TEXT PRIMARY KEY,
			topic_id   TEXT NOT NULL,
			user_id    TEXT NOT NULL,
			word       TEXT NOT NULL,
			pos        TEXT NOT NULL DEFAULT 'other',
			description TEXT NOT NULL DEFAULT '',
			example    TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
		)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatal("Failed to create table:", err)
		}
	}

	// Enable foreign keys
	DB.Exec("PRAGMA foreign_keys = ON")
}
