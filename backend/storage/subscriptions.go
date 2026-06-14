package storage

import (
	"database/sql"
	"subhub-backend/structs"
)

func GetSubscriptions(db *sql.DB, userID int64,
	category string) ([]structs.Subscription, error) {
	if err := normalizeUserSubscriptionNextPayments(db, userID); err != nil {
		return nil, err
	}

	query := `
SELECT id, user_id, name, price, period, category, next_payment, link, status, comment, plan_type
FROM subscriptions
WHERE user_id = ?`
	args := []any{userID}

	if category != "" && category != "Все" {
		query += " AND category = ?"
		args = append(args, category)
	}

	query += " ORDER BY next_payment ASC, id DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	subs := make([]structs.Subscription, 0)
	for rows.Next() {
		var sub structs.Subscription
		if err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.Name,
			&sub.Price,
			&sub.Period,
			&sub.Category,
			&sub.NextPayment,
			&sub.Link,
			&sub.Status,
			&sub.Comment,
			&sub.PlanType,
		); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

func AddSubscription(db *sql.DB, sub structs.Subscription) (int64, error) {
	result, err := db.Exec(`
INSERT INTO subscriptions (
    user_id, name, price, period, category, next_payment, link, status, comment, plan_type
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sub.UserID,
		sub.Name,
		sub.Price,
		sub.Period,
		sub.Category,
		sub.NextPayment,
		sub.Link,
		sub.Status,
		sub.Comment,
		sub.PlanType,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	sub.ID = id
	if err := syncRecentPaymentHistory(db, sub); err != nil {
		return 0, err
	}

	return id, nil
}

func UpdateSubscription(db *sql.DB, sub structs.Subscription) error {
	result, err := db.Exec(`
UPDATE subscriptions
SET name = ?, price = ?, period = ?, category = ?, 
    next_payment = ?, link = ?, status = ?, comment = ?, plan_type = ?
WHERE id = ? AND user_id = ?`,
		sub.Name,
		sub.Price,
		sub.Period,
		sub.Category,
		sub.NextPayment,
		sub.Link,
		sub.Status,
		sub.Comment,
		sub.PlanType,
		sub.ID,
		sub.UserID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if err := syncRecentPaymentHistory(db, sub); err != nil {
		return err
	}
	if err := syncGroupTotalsBySubscription(db, sub.ID); err != nil {
		return err
	}

	return nil
}

func DeleteSubscription(db *sql.DB, id int64, userID int64) error {
	var exists int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM subscriptions WHERE id = ? AND user_id = ?",
		id,
		userID,
	).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return sql.ErrNoRows
	}

	groupIDs, err := getLinkedGroupIDs(db, id)
	if err != nil {
		return err
	}
	if _, err := db.Exec("DELETE FROM group_subscriptions WHERE subscription_id = ?",
		id); err != nil {
		return err
	}

	result, err := db.Exec("DELETE FROM subscriptions WHERE id = ? AND user_id = ?",
		id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	for _, groupID := range groupIDs {
		if err := recalculateGroupTotal(db, groupID); err != nil {
			return err
		}
	}
	if err := syncSubscriptionPlanTypes(db, []int64{id}); err != nil {
		return err
	}

	return nil
}
