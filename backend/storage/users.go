package storage

import (
	"database/sql"
	"errors"

	"subhub-backend/structs"
)

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
