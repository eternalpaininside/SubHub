package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"subhub-backend/storage"
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
