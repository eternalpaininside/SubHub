package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"subhub-backend/storage"
	"subhub-backend/structs"
)

func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type",
		"application/json")
	w.Header().Set("Access-Control-Allow-Origin",
		"*")
	w.Header().Set("Access-Control-Allow-Methods",
		"GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers",
		"Content-Type")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	setCommonHeaders(w)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w,
		status,
		map[string]string{"error": message},
	)
}

func prepareAPIRequest(w http.ResponseWriter, r *http.Request) bool {
	setCommonHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return false
	}
	return true
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		writeError(w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return false
	}
	return true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w,
			http.StatusBadRequest,
			"invalid json",
		)
		return false
	}
	return true
}

func writeNotFoundOrError(w http.ResponseWriter, err error,
	notFoundMessage string, status int, message string) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w,
			http.StatusNotFound,
			notFoundMessage,
		)
		return
	}
	writeError(w, status, message)
}

func parseUserID(value string) (int64, error) {
	userID, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || userID <= 0 {
		return 0, errors.New("invalid user_id")
	}
	return userID, nil
}

func readQueryID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := parseUserID(r.URL.Query().Get(name))
	if err != nil {
		writeError(w,
			http.StatusBadRequest,
			err.Error(),
		)
		return 0, false
	}
	return id, true
}

func extractIDFromPath(path string, prefix string) (int64, error) {
	rawID := strings.TrimPrefix(path, prefix)
	rawID = strings.Trim(rawID, "/")
	return strconv.ParseInt(rawID, 10, 64)
}

func AuthRegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !prepareAPIRequest(w, r) ||
			!requireMethod(w, r, http.MethodPost) {
			return
		}

		var req structs.AuthRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		req.Password = strings.TrimSpace(req.Password)

		if req.Name == "" || req.Email == "" || req.Password == "" {
			writeError(w,
				http.StatusBadRequest,
				"name, email and password are required",
			)
			return
		}
		if len(req.Password) < 6 {
			writeError(w,
				http.StatusBadRequest,
				"password must be at least 6 characters",
			)
			return
		}

		user, err := storage.CreateUser(db, req)
		if err != nil {
			log.Printf("register user failed for email=%q: %v",
				req.Email, err)
			if strings.Contains(strings.ToLower(err.Error()),
				"unique constraint failed: users.email") {
				writeError(w,
					http.StatusConflict,
					"user with this email already exists",
				)
				return
			}
			writeError(w,
				http.StatusBadRequest,
				err.Error(),
			)
			return
		}

		writeJSON(w,
			http.StatusCreated,
			map[string]any{"user": user},
		)
	}
}

func AuthLoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !prepareAPIRequest(w, r) ||
			!requireMethod(w, r, http.MethodPost) {
			return
		}

		var req structs.AuthRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if strings.TrimSpace(req.Email) == "" ||
			strings.TrimSpace(req.Password) == "" {
			writeError(w,
				http.StatusBadRequest,
				"email and password are required",
			)
			return
		}

		user, err := storage.AuthenticateUser(db,
			strings.TrimSpace(strings.ToLower(req.Email)),
			req.Password)
		if err != nil {
			writeError(w,
				http.StatusUnauthorized,
				"invalid email or password",
			)
			return
		}

		writeJSON(w,
			http.StatusOK,
			map[string]any{"user": user},
		)
	}
}

func SubscriptionsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !prepareAPIRequest(w, r) {
			return
		}

		switch r.Method {
		case http.MethodGet:
			userID, ok := readQueryID(w, r, "user_id")
			if !ok {
				return
			}

			subs, err := storage.GetSubscriptions(db, userID,
				r.URL.Query().Get("category"))
			if err != nil {
				writeError(w,
					http.StatusInternalServerError,
					"failed to load subscriptions",
				)
				return
			}

			writeJSON(w, http.StatusOK, subs)
		case http.MethodPost:
			var sub structs.Subscription
			if !decodeJSON(w, r, &sub) {
				return
			}
			if sub.UserID <= 0 ||
				strings.TrimSpace(sub.Name) == "" ||
				strings.TrimSpace(sub.NextPayment) == "" {
				writeError(w,
					http.StatusBadRequest,
					"invalid subscription payload",
				)
				return
			}
			if sub.PlanType == "" {
				sub.PlanType = "Индивидуальный"
			}

			id, err := storage.AddSubscription(db, sub)
			if err != nil {
				writeError(w,
					http.StatusBadRequest,
					"failed to add subscription",
				)
				return
			}
			sub.ID = id

			writeJSON(w,
				http.StatusCreated,
				sub)
		default:
			writeError(w,
				http.StatusMethodNotAllowed,
				"method not allowed",
			)
		}
	}
}

func SubscriptionByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !prepareAPIRequest(w, r) {
			return
		}

		id, err := extractIDFromPath(r.URL.Path,
			"/api/subscriptions/")
		if err != nil {
			writeError(w,
				http.StatusBadRequest,
				"invalid subscription id",
			)
			return
		}

		switch r.Method {
		case http.MethodPut:
			var sub structs.Subscription
			if !decodeJSON(w, r, &sub) {
				return
			}
			if sub.UserID <= 0 {
				writeError(w,
					http.StatusBadRequest,
					"invalid user_id",
				)
				return
			}
			sub.ID = id

			if sub.PlanType == "" {
				sub.PlanType = "Индивидуальный"
			}

			if err := storage.UpdateSubscription(db, sub); err != nil {
				writeNotFoundOrError(w,
					err,
					"subscription not found",
					http.StatusBadRequest,
					"failed to update subscription",
				)
				return
			}

			writeJSON(w, http.StatusOK, sub)
		case http.MethodDelete:
			userID, ok := readQueryID(w, r, "user_id")
			if !ok {
				return
			}

			if err := storage.DeleteSubscription(db, id, userID); err != nil {
				writeNotFoundOrError(w,
					err,
					"subscription not found",
					http.StatusInternalServerError,
					"failed to delete subscription",
				)
				return
			}

			writeJSON(w,
				http.StatusOK,
				map[string]any{"deleted": true},
			)
		default:
			writeError(w,
				http.StatusMethodNotAllowed,
				"method not allowed",
			)
		}
	}
}

func AnalyticsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !prepareAPIRequest(w, r) ||
			!requireMethod(w, r, http.MethodGet) {
			return
		}

		userID, ok := readQueryID(w, r, "user_id")
		if !ok {
			return
		}

		analytics, err := storage.GetAnalytics(db, userID)
		if err != nil {
			writeError(w,
				http.StatusInternalServerError,
				"failed to load analytics",
			)
			return
		}

		writeJSON(w,
			http.StatusOK,
			analytics,
		)
	}
}

func ProfileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !prepareAPIRequest(w, r) ||
			!requireMethod(w, r, http.MethodGet) {
			return
		}

		userID, ok := readQueryID(w, r, "user_id")
		if !ok {
			return
		}

		profile, err := storage.GetProfile(db, userID)
		if err != nil {
			writeNotFoundOrError(w,
				err,
				"user not found",
				http.StatusInternalServerError,
				"failed to load profile",
			)
			return
		}

		writeJSON(w,
			http.StatusOK,
			profile,
		)
	}
}

func GroupsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !prepareAPIRequest(w, r) {
			return
		}

		switch r.Method {
		case http.MethodGet:
			userID, ok := readQueryID(w, r, "user_id")
			if !ok {
				return
			}

			groups, err := storage.GetGroups(db, userID)
			if err != nil {
				writeError(w,
					http.StatusInternalServerError,
					"failed to load groups",
				)
				return
			}

			writeJSON(w,
				http.StatusOK,
				groups,
			)
		case http.MethodPost:
			var group structs.Group
			if !decodeJSON(w, r, &group) {
				return
			}
			if group.OwnerID <= 0 ||
				strings.TrimSpace(group.Name) == "" {
				writeError(w,
					http.StatusBadRequest,
					"invalid group payload",
				)
				return
			}
			if group.Type == "" {
				group.Type = "Семейная"
			}

			id, inviteURL, err := storage.CreateGroup(db, group)
			if err != nil {
				writeError(w,
					http.StatusBadRequest,
					"failed to create group",
				)
				return
			}
			group.ID = id
			group.InviteURL = inviteURL

			writeJSON(w,
				http.StatusCreated,
				group,
			)
		default:
			writeError(w,
				http.StatusMethodNotAllowed,
				"method not allowed",
			)
		}
	}
}

func GroupByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !prepareAPIRequest(w, r) {
			return
		}

		id, err := extractIDFromPath(r.URL.Path,
			"/api/groups/")
		if err != nil {
			writeError(w,
				http.StatusBadRequest,
				"invalid group id",
			)
			return
		}

		switch r.Method {
		case http.MethodPut:
			var group structs.Group
			if !decodeJSON(w, r, &group) {
				return
			}
			if group.OwnerID <= 0 ||
				strings.TrimSpace(group.Name) == "" {
				writeError(w,
					http.StatusBadRequest,
					"invalid group payload",
				)
				return
			}
			if group.Type == "" {
				group.Type = "Семейная"
			}
			group.ID = id

			if err := storage.UpdateGroup(db, group); err != nil {
				writeNotFoundOrError(w,
					err,
					"group not found",
					http.StatusBadRequest,
					"failed to update group",
				)
				return
			}

			writeJSON(w, http.StatusOK, group)
		case http.MethodDelete:
			ownerID, ok := readQueryID(w, r, "owner_id")
			if !ok {
				return
			}

			if err := storage.DeleteGroup(db, id, ownerID); err != nil {
				writeNotFoundOrError(w,
					err,
					"group not found",
					http.StatusInternalServerError,
					"failed to delete group",
				)
				return
			}

			writeJSON(w,
				http.StatusOK,
				map[string]any{"deleted": true},
			)
		default:
			writeError(w,
				http.StatusMethodNotAllowed,
				"method not allowed",
			)
		}
	}
}

func JoinGroupHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !prepareAPIRequest(w, r) ||
			!requireMethod(w, r, http.MethodPost) {
			return
		}

		var req structs.JoinGroupRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.UserID <= 0 ||
			strings.TrimSpace(req.InviteURL) == "" {
			writeError(w,
				http.StatusBadRequest,
				"user_id and invite_url are required",
			)
			return
		}

		groupID, err := storage.JoinGroupByInviteURL(db,
			req.UserID, req.InviteURL)
		if err != nil {
			writeNotFoundOrError(w,
				err,
				"group not found",
				http.StatusBadRequest,
				"failed to join group",
			)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"joined":   true,
			"group_id": groupID,
		})
	}
}
