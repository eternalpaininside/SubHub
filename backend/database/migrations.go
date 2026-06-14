package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"subhub-backend/structs"
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

func GetTime() (time.Time, time.Time) {
	windowStart := time.Now().AddDate(0, -5, 0)
	windowStart = time.Date(windowStart.Year(), windowStart.Month(),
		1, 0, 0, 0, 0, windowStart.Location())
	now := time.Now()
	return windowStart, now
}

func backfillPaymentHistory(db *sql.DB) {
	rows, err := db.Query(`
SELECT id, user_id, price, period, next_payment, status
FROM subscriptions`)
	if err != nil {
		log.Fatal(err)
	}

	windowStart, now := GetTime()
	seeds := make([]structs.PaymentHistory, 0)

	for rows.Next() {
		var seed structs.PaymentHistory
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
