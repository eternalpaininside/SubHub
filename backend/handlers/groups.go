package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"subhub-backend/storage"
	"subhub-backend/structs"
)

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
			"/groups/")
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
