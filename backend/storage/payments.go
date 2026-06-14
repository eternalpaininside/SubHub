package storage

import (
	"database/sql"
	"strings"
	"subhub-backend/database"
	"subhub-backend/structs"
	"time"
)

func normalizeUserSubscriptionNextPayments(db *sql.DB, userID int64) error {
	rows, err := db.Query(`
SELECT id, next_payment, period
FROM subscriptions
WHERE user_id = ? AND status = 1`, userID)
	if err != nil {
		return err
	}
	defer closeRows(rows)

	type paymentUpdate struct {
		id          int64
		nextPayment string
	}

	updates := make([]paymentUpdate, 0)
	now := time.Now()

	for rows.Next() {
		var (
			id          int64
			nextPayment string
			period      int64
		)
		if err := rows.Scan(&id, &nextPayment, &period); err != nil {
			return err
		}

		normalized, changed := normalizeNextPaymentDate(nextPayment, period, now)
		if changed {
			updates = append(updates,
				paymentUpdate{id: id, nextPayment: normalized})
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for _, update := range updates {
		if _, err := db.Exec(
			"UPDATE subscriptions SET next_payment = ? WHERE id = ?",
			update.nextPayment,
			update.id,
		); err != nil {
			return err
		}
	}

	return nil
}

func normalizeNextPaymentDate(raw string, period int64,
	now time.Time) (string, bool) {
	nextPayment, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return raw, false
	}

	stepMonths := int(period)
	if stepMonths <= 0 {
		stepMonths = 1
	}

	todayStart := time.Date(now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, now.Location())
	normalized := nextPayment

	for normalized.Before(todayStart) {
		normalized = normalized.AddDate(0, stepMonths, 0)
	}

	normalizedRaw := normalized.Format("2006-01-02")
	return normalizedRaw, normalizedRaw != raw
}

func syncRecentPaymentHistory(db *sql.DB, sub structs.Subscription) error {
	if !sub.Status || strings.TrimSpace(sub.NextPayment) == "" ||
		sub.ID <= 0 {
		return nil
	}

	nextPayment, err := time.Parse("2006-01-02", sub.NextPayment)
	if err != nil {
		return nil
	}

	windowStart, now := database.GetTime()

	stepMonths := int(sub.Period)
	if stepMonths <= 0 {
		stepMonths = 1
	}

	for candidate := nextPayment; !candidate.Before(windowStart); candidate = candidate.AddDate(0, -stepMonths, 0) {
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
			sub.UserID,
			sub.ID,
			sub.Price,
			candidate.Format("2006-01-02"),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncSubscriptionPlanTypes(db *sql.DB, subscriptionIDs []int64) error {
	seen := make(map[int64]struct{})
	for _, subscriptionID := range subscriptionIDs {
		if subscriptionID <= 0 {
			continue
		}
		if _, exists := seen[subscriptionID]; exists {
			continue
		}
		seen[subscriptionID] = struct{}{}

		var groupCount int
		if err := db.QueryRow(
			"SELECT COUNT(*) FROM group_subscriptions WHERE subscription_id = ?",
			subscriptionID,
		).Scan(&groupCount); err != nil {
			return err
		}

		if groupCount > 0 {
			if _, err := db.Exec(
				"UPDATE subscriptions SET plan_type = 'Групповой' WHERE id = ?",
				subscriptionID,
			); err != nil {
				return err
			}
			continue
		}

		if _, err := db.Exec(
			"UPDATE subscriptions SET plan_type = 'Индивидуальный' WHERE id = ? AND plan_type = 'Групповой'",
			subscriptionID,
		); err != nil {
			return err
		}
	}

	return nil
}

func getPaymentHistoryBars(db *sql.DB, userID int64) ([]structs.MonthAmount, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1,
		0, 0, 0, 0, now.Location()).AddDate(0, -5, 0)

	rows, err := db.Query(`
SELECT strftime('%Y-%m', paid_at) AS month_key, COALESCE(SUM(amount), 0) AS total
FROM payment_history
WHERE user_id = ?
  AND date(paid_at) BETWEEN date(?) AND date('now')
GROUP BY month_key
ORDER BY month_key ASC`,
		userID,
		start.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	totalsByMonth := make(map[string]int64)
	for rows.Next() {
		var monthKey string
		var total int64
		if err := rows.Scan(&monthKey, &total); err != nil {
			return nil, err
		}
		totalsByMonth[monthKey] = total
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	bars := make([]structs.MonthAmount, 0, 6)
	for i := 0; i < 6; i++ {
		monthDate := start.AddDate(0, i, 0)
		key := monthDate.Format("2006-01")

		bars = append(bars, structs.MonthAmount{
			Label:  localizeMonthLabel(monthDate),
			Amount: totalsByMonth[key],
		})
	}

	return bars, nil
}

func localizeMonthLabel(date time.Time) string {
	if label, ok := monthLabels[date.Month()]; ok {
		return label
	}

	return date.Format("Jan")
}
