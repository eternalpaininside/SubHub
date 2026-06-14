package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"subhub-backend/structs"
)

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
