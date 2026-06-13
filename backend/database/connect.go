package database

import (
	"database/sql"
	"log"
	"os"
)

func ConnectionDB() *sql.DB {
	dbPath := "./subscriptions.sqlite"
	if _, err := os.Stat("../subscriptions.sqlite"); err == nil {
		dbPath = "../subscriptions.sqlite"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		log.Fatal(err)
	}

	return db
}
