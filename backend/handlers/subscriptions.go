package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"subhub-backend/storage"
	"subhub-backend/structs"
)

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
			"/subscriptions/")
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
