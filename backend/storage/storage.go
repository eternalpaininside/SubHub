package storage

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"subhub-backend/database"
	"subhub-backend/structs"

	"database/sql"
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

func CreateUser(db *sql.DB, req structs.AuthRequest) (structs.User, error) {
	result, err := db.Exec(
		"INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
		req.Name,
		req.Email,
		req.Password,
	)
	if err != nil {
		return structs.User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return structs.User{}, err
	}

	return GetUserByID(db, id)
}

func AuthenticateUser(db *sql.DB, email string,
	password string) (structs.User, error) {
	user, err := GetUserByEmail(db, email)
	if err != nil {
		return structs.User{}, err
	}

	if user.Password != password {
		return structs.User{},
			errors.New("invalid credentials")
	}

	return user, nil
}

func GetUserByEmail(db *sql.DB, email string) (structs.User, error) {
	var user structs.User
	err := db.QueryRow(
		"SELECT id, name, email, password, created_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt)

	return user, err
}

func GetUserByID(db *sql.DB, id int64) (structs.User, error) {
	var user structs.User
	err := db.QueryRow(
		"SELECT id, name, email, password, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt)

	return user, err
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

func CreateGroup(db *sql.DB, group structs.Group) (int64, string, error) {
	if strings.TrimSpace(group.InviteURL) == "" {
		group.InviteURL = generateGroupInviteURL(group.OwnerID, group.Name)
	}

	result, err := db.Exec(`
INSERT INTO groups (owner_id, name, type, price, period, invite_url, notes)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		group.OwnerID,
		group.Name,
		group.Type,
		0,
		1,
		group.InviteURL,
		group.Notes,
	)
	if err != nil {
		return 0, "", err
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		return 0, "", err
	}

	_, err = db.Exec(
		"INSERT OR IGNORE INTO group_members (group_id, user_id, role) VALUES (?, ?, 'owner')",
		groupID,
		group.OwnerID,
	)
	if err != nil {
		return 0, "", err
	}

	if err := replaceGroupSubscriptions(db, groupID,
		group.OwnerID, group.SubscriptionIDs); err != nil {
		return 0, "", err
	}

	return groupID, group.InviteURL, nil
}

func UpdateGroup(db *sql.DB, group structs.Group) error {
	result, err := db.Exec(`
UPDATE groups
SET name = ?, type = ?, notes = ?
WHERE id = ? AND owner_id = ?`,
		group.Name,
		group.Type,
		group.Notes,
		group.ID,
		group.OwnerID,
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

	return replaceGroupSubscriptions(db, group.ID, group.OwnerID, group.SubscriptionIDs)
}

func DeleteGroup(db *sql.DB, groupID int64, ownerID int64) error {
	linkedSubscriptionIDs, err := getGroupSubscriptionIDs(db, groupID)
	if err != nil {
		return err
	}

	var exists int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM groups WHERE id = ? AND owner_id = ?",
		groupID,
		ownerID,
	).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return sql.ErrNoRows
	}

	if _, err := db.Exec("DELETE FROM group_subscriptions WHERE group_id = ?",
		groupID); err != nil {
		return err
	}
	if _, err := db.Exec("DELETE FROM group_members WHERE group_id = ?",
		groupID); err != nil {
		return err
	}

	result, err := db.Exec("DELETE FROM groups WHERE id = ? AND owner_id = ?",
		groupID, ownerID)
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
	if err := syncSubscriptionPlanTypes(db, linkedSubscriptionIDs); err != nil {
		return err
	}
	return nil
}

func JoinGroupByInviteURL(db *sql.DB, userID int64, inviteURL string) (int64, error) {
	inviteURL = strings.TrimSpace(inviteURL)
	if inviteURL == "" {
		return 0, errors.New("invite_url is required")
	}

	var groupID int64
	err := db.QueryRow(
		"SELECT id FROM groups WHERE invite_url = ?",
		inviteURL,
	).Scan(&groupID)
	if err != nil {
		return 0, err
	}

	_, err = db.Exec(
		"INSERT OR IGNORE INTO group_members (group_id, user_id, role) VALUES (?, ?, 'member')",
		groupID,
		userID,
	)
	if err != nil {
		return 0, err
	}

	return groupID, nil
}

func GetGroups(db *sql.DB, userID int64) ([]structs.Group, error) {
	type groupRow struct {
		ID        int64
		OwnerID   int64
		Name      string
		Type      string
		Period    int64
		InviteURL string
		Notes     string
	}

	rows, err := db.Query(`
SELECT g.id, g.owner_id, g.name, g.type, g.period, g.invite_url, g.notes
FROM groups g
LEFT JOIN group_members gm ON gm.group_id = g.id
WHERE g.owner_id = ? OR gm.user_id = ?
GROUP BY g.id
ORDER BY g.id DESC`,
		userID,
		userID,
	)
	if err != nil {
		return nil, err
	}

	groupRows := make([]groupRow, 0)
	for rows.Next() {
		var item groupRow
		if err := rows.Scan(
			&item.ID,
			&item.OwnerID,
			&item.Name,
			&item.Type,
			&item.Period,
			&item.InviteURL,
			&item.Notes,
		); err != nil {
			closeRows(rows)
			return nil, err
		}
		groupRows = append(groupRows, item)
	}

	groups := make([]structs.Group, 0, len(groupRows))
	for _, item := range groupRows {
		members, err := getGroupMembers(db, item.ID, item.OwnerID)
		if err != nil {
			return nil, err
		}
		services, subscriptionIDs, monthlyTotal, err := getGroupServices(db,
			item.ID, item.OwnerID)
		if err != nil {
			return nil, err
		}
		groups = append(groups, structs.Group{
			ID:              item.ID,
			OwnerID:         item.OwnerID,
			Name:            item.Name,
			Type:            item.Type,
			Price:           monthlyTotal,
			Period:          1,
			InviteURL:       item.InviteURL,
			Notes:           item.Notes,
			Members:         members,
			Services:        services,
			SubscriptionIDs: subscriptionIDs,
		})
	}

	return groups, nil
}

func getGroupMembers(db *sql.DB, groupID int64, ownerID int64) ([]structs.GroupMember, error) {
	rows, err := db.Query(`
SELECT u.id, u.name
FROM group_members gm
JOIN users u ON u.id = gm.user_id
WHERE gm.group_id = ?
ORDER BY CASE WHEN gm.role = 'owner' THEN 0 ELSE 1 END, u.name`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	members := make([]structs.GroupMember, 0)
	for rows.Next() {
		var member structs.GroupMember
		if err := rows.Scan(&member.UserID, &member.Name); err != nil {
			return nil, err
		}
		member.Owner = member.UserID == ownerID
		members = append(members, member)
	}

	return members, rows.Err()
}

func generateGroupInviteURL(ownerID int64, groupName string) string {
	slug := strings.ToLower(strings.TrimSpace(groupName))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = regexp.MustCompile(`[^a-z0-9\-_а-я]+`).ReplaceAllString(slug, "")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "group"
	}

	return fmt.Sprintf("https://subhub.local/invite/%d-%s", ownerID, slug)
}

func replaceGroupSubscriptions(db *sql.DB, groupID int64,
	ownerID int64, subscriptionIDs []int64) error {
	previousSubscriptionIDs, err := getGroupSubscriptionIDs(db, groupID)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(""+
		"DELETE FROM group_subscriptions WHERE group_id = ?",
		groupID); err != nil {
		return err
	}

	seen := make(map[int64]struct{})
	for _, subscriptionID := range subscriptionIDs {
		if subscriptionID <= 0 {
			continue
		}
		if _, exists := seen[subscriptionID]; exists {
			continue
		}
		seen[subscriptionID] = struct{}{}

		var exists int
		if err := tx.QueryRow(
			"SELECT COUNT(*) FROM subscriptions WHERE id = ? AND user_id = ?",
			subscriptionID,
			ownerID,
		).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			continue
		}

		if _, err := tx.Exec(
			"INSERT INTO group_subscriptions (group_id, subscription_id) VALUES (?, ?)",
			groupID,
			subscriptionID,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if err := recalculateGroupTotal(db, groupID); err != nil {
		return err
	}

	return syncSubscriptionPlanTypes(db,
		append(previousSubscriptionIDs, subscriptionIDs...))
}

func getGroupServices(db *sql.DB, groupID int64, ownerID int64) ([]string, []int64, int64, error) {
	rows, err := db.Query(`
SELECT s.id, s.name, CASE WHEN s.period = 12 THEN s.price / 12 ELSE s.price END AS monthly_price
FROM group_subscriptions gs
JOIN subscriptions s ON s.id = gs.subscription_id
WHERE gs.group_id = ? AND s.user_id = ?
ORDER BY s.name`,
		groupID,
		ownerID,
	)
	if err != nil {
		return nil, nil, 0, err
	}
	defer closeRows(rows)

	services := make([]string, 0)
	subscriptionIDs := make([]int64, 0)
	var monthlyTotal int64

	for rows.Next() {
		var (
			subscriptionID int64
			name           string
			monthlyPrice   int64
		)
		if err := rows.Scan(&subscriptionID, &name, &monthlyPrice); err != nil {
			return nil, nil, 0, err
		}
		subscriptionIDs = append(subscriptionIDs, subscriptionID)
		services = append(services, name)
		monthlyTotal += monthlyPrice
	}

	return services, subscriptionIDs, monthlyTotal, rows.Err()
}

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

func getIdsFromRows(rows *sql.Rows) ([]int64, error) {
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func getLinkedGroupIDs(db *sql.DB, subscriptionID int64) ([]int64, error) {
	rows, err := db.Query(
		"SELECT group_id FROM group_subscriptions WHERE subscription_id = ?",
		subscriptionID)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	groupIDs, err := getIdsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return groupIDs, rows.Err()
}

func getGroupSubscriptionIDs(db *sql.DB, groupID int64) ([]int64, error) {
	rows, err := db.Query(
		"SELECT subscription_id FROM group_subscriptions WHERE group_id = ?",
		groupID)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	subscriptionIDs, err := getIdsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return subscriptionIDs, rows.Err()
}

func recalculateGroupTotal(db *sql.DB, groupID int64) error {
	var monthlyTotal int64
	if err := db.QueryRow(`
SELECT COALESCE(SUM(CASE WHEN s.period = 12 THEN s.price / 12 ELSE s.price END), 0)
FROM group_subscriptions gs
JOIN subscriptions s ON s.id = gs.subscription_id
WHERE gs.group_id = ?`,
		groupID,
	).Scan(&monthlyTotal); err != nil {
		return err
	}

	_, err := db.Exec("UPDATE groups SET price = ?, period = 1 WHERE id = ?",
		monthlyTotal, groupID)
	return err
}

func syncGroupTotalsBySubscription(db *sql.DB, subscriptionID int64) error {
	groupIDs, err := getLinkedGroupIDs(db, subscriptionID)
	if err != nil {
		return err
	}

	for _, groupID := range groupIDs {
		if err := recalculateGroupTotal(db, groupID); err != nil {
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
