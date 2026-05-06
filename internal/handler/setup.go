package handler

import (
	"net/http"

	"mo/internal/database"
	"mo/internal/model"
	"mo/internal/service"
	"mo/internal/config"

	"github.com/gin-gonic/gin"
)

type SetupHandler struct {
	JWTSecret string
	Config    *config.Config
}

func NewSetupHandler(secret string, cfg *config.Config) *SetupHandler {
	return &SetupHandler{JWTSecret: secret, Config: cfg}
}

func (h *SetupHandler) Status(c *gin.Context) {
	setup, err := database.IsSetup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Database error"},
		})
		return
	}
	c.JSON(http.StatusOK, model.SetupStatusResponse{SetupRequired: !setup})
}

func (h *SetupHandler) Initialize(c *gin.Context) {
	setup, err := database.IsSetup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Database error"},
		})
		return
	}
	if setup {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{"code": "ALREADY_SETUP", "message": "System is already initialized"},
		})
		return
	}

	var req model.SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "All required fields must be provided"},
		})
		return
	}

	// Validate password strength
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Password must be at least 8 characters"},
		})
		return
	}

	// Create admin user
	resp, err := service.CreateAdmin(req.Username, req.Email, req.Password, h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Could not create admin user: " + err.Error()},
		})
		return
	}

	// Save site settings
	h.Config.Site.Title = req.SiteTitle
	h.Config.Site.Subtitle = req.SiteSubtitle
	h.Config.Site.Description = req.SiteDesc

	// Insert settings into database
	settings := map[string]string{
		"site.title":       req.SiteTitle,
		"site.subtitle":    req.SiteSubtitle,
		"site.description": req.SiteDesc,
	}
	for k, v := range settings {
		if _, err := database.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", k, v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL_ERROR", "message": "Could not save settings"},
			})
			return
		}
	}

	// Set refresh token cookie
	refreshToken, _ := generateRefreshToken(resp.User.ID, resp.User.Username, h.JWTSecret)
	if refreshToken != "" {
		c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/api/v1/auth", "", false, true)
	}

	c.JSON(http.StatusOK, resp)
}
