package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"subhub-backend/storage"
	"subhub-backend/structs"
)

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
