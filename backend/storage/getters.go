package storage

import (
	"database/sql"
	"log"
	"time"

	"subhub-backend/structs"
)

var monthLabels = map[time.Month]string{
	time.January:   "Янв",
	time.February:  "Фев",
	time.March:     "Мар",
	time.April:     "Апр",
	time.May:       "Май",
	time.June:      "Июн",
	time.July:      "Июл",
	time.August:    "Авг",
	time.September: "Сен",
	time.October:   "Окт",
	time.November:  "Ноя",
	time.December:  "Дек",
}

func closeRows(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func GetAnalytics(db *sql.DB, userID int64) (structs.AnalyticsResponse, error) {
	var analytics structs.AnalyticsResponse

	if err := normalizeUserSubscriptionNextPayments(db, userID); err != nil {
		return analytics, err
	}

	if err := db.QueryRow(
		"SELECT COUNT(*) FROM subscriptions WHERE user_id = ? AND status = 1",
		userID,
	).Scan(&analytics.ActiveCount); err != nil {
		return analytics, err
	}

	if err := db.QueryRow(`
SELECT COUNT(*)
FROM subscriptions
WHERE user_id = ?
  AND status = 1
  AND date(next_payment) BETWEEN date('now') AND date('now', '+7 day')`,
		userID,
	).Scan(&analytics.ExpiringCount); err != nil {
		return analytics, err
	}

	if err := db.QueryRow(`
SELECT COALESCE(SUM(CASE WHEN period = 12 THEN price / 12 ELSE price END), 0)
FROM subscriptions
WHERE user_id = ? AND status = 1`,
		userID,
	).Scan(&analytics.MonthlyTotal); err != nil {
		return analytics, err
	}

	analytics.YearlyEstimate = analytics.MonthlyTotal * 12

	bars, err := getPaymentHistoryBars(db, userID)
	if err != nil {
		return analytics, err
	}
	analytics.Bars = bars

	rows, err := db.Query(`
SELECT category, COALESCE(SUM(CASE WHEN period = 12 THEN price / 12 ELSE price END), 0) AS total
FROM subscriptions
WHERE user_id = ? AND status = 1
GROUP BY category
ORDER BY total DESC, category ASC`,
		userID,
	)
	if err != nil {
		return analytics, err
	}
	defer closeRows(rows)

	for rows.Next() {
		var item structs.CategorySummary
		if err := rows.Scan(&item.Category, &item.Total); err != nil {
			return analytics, err
		}
		analytics.ByCategory = append(analytics.ByCategory, item)
	}

	return analytics, rows.Err()
}

func GetProfile(db *sql.DB, userID int64) (structs.ProfileResponse, error) {
	user, err := GetUserByID(db, userID)
	if err != nil {
		return structs.ProfileResponse{}, err
	}

	analytics, err := GetAnalytics(db, userID)
	if err != nil {
		return structs.ProfileResponse{}, err
	}

	var groupCount int64
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM groups WHERE owner_id = ?",
		userID,
	).Scan(&groupCount); err != nil {
		return structs.ProfileResponse{}, err
	}

	return structs.ProfileResponse{
		User: user,
		Stats: structs.ProfileStats{
			ActiveSubscriptions: analytics.ActiveCount,
			GroupCount:          groupCount,
			MonthlySpend:        analytics.MonthlyTotal,
		},
	}, nil
}
