package service

import (
	"database/sql"
	"errors"

	"mo/internal/auth"
	"mo/internal/database"
	"mo/internal/model"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

func Login(login, password, jwtSecret string) (*model.LoginResponse, error) {
	var user model.User
	var passwordHash string
	err := database.DB.QueryRow(
		"SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = ? OR email = ?",
		login, login,
	).Scan(&user.ID, &user.Username, &user.Email, &passwordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Username, jwtSecret)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken: accessToken,
		User:        user,
	}, nil
}

func CreateAdmin(username, email, password, jwtSecret string) (*model.LoginResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	user := model.User{
		ID:           model.NewULID(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}

	_, err = database.DB.Exec(
		"INSERT INTO users (id, username, email, password_hash) VALUES (?, ?, ?, ?)",
		user.ID, user.Username, user.Email, user.PasswordHash,
	)
	if err != nil {
		return nil, err
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Username, jwtSecret)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken: accessToken,
		User:        user,
	}, nil
}

func RefreshToken(userID, username, secret string) (string, error) {
	return auth.GenerateAccessToken(userID, username, secret)
}

func ChangePassword(userID, oldPassword, newPassword string) error {
	var currentHash string
	err := database.DB.QueryRow("SELECT password_hash FROM users WHERE id = ?", userID).Scan(&currentHash)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec("UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", string(newHash), userID)
	return err
}
