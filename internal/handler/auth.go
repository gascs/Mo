package handler

import (
	"net/http"

	"mo/internal/model"
	"mo/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	JWTSecret string
}

func NewAuthHandler(secret string) *AuthHandler {
	return &AuthHandler{JWTSecret: secret}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Login and password are required"},
		})
		return
	}

	resp, err := service.Login(req.Login, req.Password, h.JWTSecret)
	if err != nil {
		status := http.StatusInternalServerError
		errCode := "INTERNAL_ERROR"
		errMsg := "An internal error occurred"

		if err == service.ErrInvalidCredentials {
			status = http.StatusUnauthorized
			errCode = "UNAUTHORIZED"
			errMsg = "Invalid credentials"
		}

		c.JSON(status, gin.H{
			"error": gin.H{"code": errCode, "message": errMsg},
		})
		return
	}

	// Set refresh token in httpOnly cookie
	refreshToken, err := generateRefreshToken(resp.User.ID, resp.User.Username, h.JWTSecret)
	if err == nil {
		c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/api/v1/auth", "", false, true)
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "No refresh token"},
		})
		return
	}

	claims, err := validateToken(refreshToken, h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid refresh token"},
		})
		return
	}

	accessToken, err := service.RefreshToken(claims.UserID, claims.Username, h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Could not generate token"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Both old and new password are required, new password must be at least 8 characters"},
		})
		return
	}

	if err := service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "Old password is incorrect"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed"})
}
