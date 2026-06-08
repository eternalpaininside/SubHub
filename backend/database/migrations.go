package database

import (
	"database/sql"
	"fmt"
	"log"
	"subhub-backend/structs"
	"time"
)

func RunMigrations(db *sql.DB) {
	createUsersTable(db)
	createSubscriptionsTable(db)
	createGroupsTable(db)
	createGroupMembersTable(db)
	createGroupSubscriptionsTable(db)
	createPaymentHistoryTable(db)
	normalizePaymentHistoryTable(db)
	backfillPaymentHistory(db)
	syncExistingGroupState(db)
}

func createUsersTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
)`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

	ensureColumn(db, "users", "password", "TEXT NOT NULL DEFAULT ''")
}

func createSubscriptionsTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    price INTEGER NOT NULL,
    period INTEGER NOT NULL,
    category TEXT NOT NULL,
    next_payment TEXT NOT NULL,
    link TEXT NOT NULL,
    status BOOLEAN NOT NULL DEFAULT 1,
    comment TEXT NOT NULL DEFAULT '',
    plan_type TEXT NOT NULL DEFAULT 'Индивидуальный',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    price INTEGER NOT NULL DEFAULT 0,
    period INTEGER NOT NULL DEFAULT 1,
    invite_url TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id)
)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func createGroupMembersTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS group_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    joined_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func createGroupSubscriptionsTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS group_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    subscription_id INTEGER NOT NULL,
    UNIQUE(group_id, subscription_id),
    FOREIGN KEY (group_id) REFERENCES groups(id),
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func createPaymentHistoryTable(db *sql.DB) {
	query := `
CREATE TABLE IF NOT EXISTS payment_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    subscription_id INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    paid_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(subscription_id, paid_at),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}

func normalizePaymentHistoryTable(db *sql.DB) {
	if !tableExists(db, "payment_history") {
		return
	}
	if !hasColumn(db, "payment_history", "subscription_name") &&
		!hasColumn(db, "payment_history", "category") {
		return
	}

	createPaymentHistoryTable(db)

	if _, err := db.Exec(`
INSERT INTO payment_history (id, user_id, subscription_id, amount, paid_at, created_at)
SELECT id, user_id, subscription_id, amount, paid_at, COALESCE(created_at, CURRENT_TIMESTAMP)
FROM payment_history`); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec("DROP TABLE payment_history"); err != nil {
		log.Fatal(err)
	}
}

func backfillPaymentHistory(db *sql.DB) {
	rows, err := db.Query(`
SELECT id, user_id, price, period, next_payment, status
FROM subscriptions`)
	if err != nil {
		log.Fatal(err)
	}

	windowStart := time.Now().AddDate(0, -5, 0)
	windowStart = time.Date(windowStart.Year(), windowStart.Month(),
		1, 0, 0, 0, 0, windowStart.Location())
	now := time.Now()
	seeds := make([]structs.PaymentHistorySeed, 0)

	for rows.Next() {
		var seed structs.PaymentHistorySeed
		if err := rows.Scan(
			&seed.SubscriptionID,
			&seed.UserID,
			&seed.Price,
			&seed.Period,
			&seed.NextPaymentRaw,
			&seed.Status,
		); err != nil {
			log.Fatal(err)
		}
		seeds = append(seeds, seed)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	if err := rows.Close(); err != nil {
		log.Fatal(err)
	}

	for _, seed := range seeds {
		if !seed.Status || seed.NextPaymentRaw == "" {
			continue
		}

		nextPayment, err := time.Parse("2006-01-02", seed.NextPaymentRaw)
		if err != nil {
			continue
		}

		for candidate := nextPayment; !candidate.Before(windowStart); candidate = candidate.AddDate(0, -int(seed.Period), 0) {
			if candidate.After(now) {
				continue
			}
			if candidate.Before(windowStart) {
				break
			}

			_, err := db.Exec(`
INSERT INTO payment_history (
    user_id, subscription_id, amount, paid_at
)
VALUES (?, ?, ?, ?)
ON CONFLICT(subscription_id, paid_at) DO UPDATE SET
    amount = excluded.amount`,
				seed.UserID,
				seed.SubscriptionID,
				seed.Price,
				candidate.Format("2006-01-02"),
			)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func syncExistingGroupState(db *sql.DB) {
	if _, err := db.Exec(`
UPDATE groups
SET price = COALESCE((
    SELECT SUM(CASE WHEN s.period = 12 THEN s.price / 12 ELSE s.price END)
    FROM group_subscriptions gs
    JOIN subscriptions s ON s.id = gs.subscription_id
    WHERE gs.group_id = groups.id
), 0),
period = 1`); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(`
UPDATE subscriptions
SET plan_type = 'Групповой'
WHERE id IN (SELECT subscription_id FROM group_subscriptions)`); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(`
UPDATE subscriptions
SET plan_type = 'Индивидуальный'
WHERE plan_type = 'Групповой'
  AND id NOT IN (SELECT subscription_id FROM group_subscriptions)`); err != nil {
		log.Fatal(err)
	}
}

func hasColumn(db *sql.DB, tableName string, columnName string) bool {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)",
		tableName))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)
		if err := rows.Scan(&cid,
			&name,
			&columnType,
			&notNull,
			&defaultVal,
			&pk); err != nil {
			log.Fatal(err)
		}
		if name == columnName {
			return true
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return false
}

func tableExists(db *sql.DB, tableName string) bool {
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?",
		tableName).Scan(&count); err != nil {
		log.Fatal(err)
	}
	return count > 0
}

func ensureColumn(db *sql.DB, tableName string, columnName string, definition string) {
	if hasColumn(db, tableName, columnName) {
		return
	}

	alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
		tableName,
		columnName,
		definition)
	if _, err := db.Exec(alterQuery); err != nil {
		log.Fatal(err)
	}
}
