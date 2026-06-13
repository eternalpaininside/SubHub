package database

import (
	"database/sql"
	"log"
)

func createUsersTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS users (
    id 		   SERIAL          PRIMARY KEY,
    name	   TEXT    NOT NULL,
    email      TEXT    NOT NULL UNIQUE,
    password   TEXT    NOT NULL,
    created_at TEXT    NOT NULL DEFAULT     CURRENT_TIMESTAMP
)`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

	ensureColumn(db, "users", "password", "TEXT NOT NULL DEFAULT ''")
}

func createSubscriptionsTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS subscriptions (
    id           SERIAL  		  PRIMARY KEY,
    user_id      INTEGER NOT NULL,
    name         TEXT 	 NOT NULL,
    price        INTEGER NOT NULL,
    period       INTEGER NOT NULL,
    category     TEXT 	 NOT NULL,
    next_payment TEXT 	 NOT NULL,
    link         TEXT 	 NOT NULL,
    status       BOOLEAN NOT NULL DEFAULT 1,
    comment 	 TEXT 	 NOT NULL DEFAULT '',
    plan_type 	 TEXT 	 NOT NULL DEFAULT 'Индивидуальный',
    created_at   TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id)
)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

	ensureColumn(db, "subscriptions",
		"comment",
		"TEXT NOT NULL DEFAULT ''")
	ensureColumn(db, "subscriptions",
		"plan_type",
		"TEXT NOT NULL DEFAULT 'Индивидуальный'")
	ensureColumn(db, "subscriptions",
		"created_at",
		"TEXT NOT NULL DEFAULT ''")
}

func createGroupsTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS groups (
    id 		   SERIAL 			PRIMARY KEY,
    owner_id   INTEGER NOT NULL,
    name 	   TEXT    NOT NULL,
    type 	   TEXT    NOT NULL,
    price 	   INTEGER NOT NULL DEFAULT 0,
    period 	   INTEGER NOT NULL DEFAULT 1,
    invite_url TEXT    NOT NULL DEFAULT '',
    notes TEXT 		   NOT NULL DEFAULT '',
    created_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (owner_id) REFERENCES users(id)
)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func createGroupMembersTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS group_members (
    id 		  SERIAL 		   PRIMARY KEY,
    group_id  INTEGER NOT NULL,
    user_id   INTEGER NOT NULL,
    role 	  TEXT	  NOT NULL DEFAULT 'member',
    joined_at TEXT 	  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(id),
    FOREIGN KEY (user_id)  REFERENCES users(id)
)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func createGroupSubscriptionsTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS group_subscriptions (
    id 				SERIAL 			 PRIMARY KEY,
    group_id 		INTEGER NOT NULL,
    subscription_id INTEGER NOT NULL,
    
    UNIQUE(group_id, subscription_id),
    FOREIGN KEY (group_id) 		  REFERENCES groups(id),
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func createPaymentHistoryTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS payment_history (
    id 				SERIAL 			 PRIMARY KEY,
    user_id 		INTEGER NOT NULL,
    subscription_id INTEGER NOT NULL,
    amount 			INTEGER NOT NULL,
    paid_at 		TEXT 	NOT NULL,
    created_at 		TEXT 	NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(subscription_id, paid_at),
    FOREIGN KEY (user_id) 		  REFERENCES users(id),
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}
