package handler

import (
	"net/http"

	"mo/internal/config"
	"mo/internal/service"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	cfg *config.Config
}

func NewSettingsHandler(cfg *config.Config) *SettingsHandler {
	return &SettingsHandler{cfg: cfg}
}

// AdminGetSettings returns all settings grouped by category.
func (h *SettingsHandler) AdminGetSettings(c *gin.Context) {
	dbSettings, err := service.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	// Merge DB settings with config defaults
	result := map[string]interface{}{
		"site": gin.H{
			"title":       or(dbSettings["site.title"], h.cfg.Site.Title),
			"subtitle":    or(dbSettings["site.subtitle"], h.cfg.Site.Subtitle),
			"description": or(dbSettings["site.description"], h.cfg.Site.Description),
			"language":    h.cfg.Site.Language,
		},
		"theme": gin.H{
			"name":         or(dbSettings["theme.name"], h.cfg.Theme.Name),
			"accent_color": or(dbSettings["theme.accent_color"], h.cfg.Theme.AccentColor),
			"font_body":    or(dbSettings["theme.font_body"], h.cfg.Theme.FontBody),
			"font_code":    or(dbSettings["theme.font_code"], h.cfg.Theme.FontCode),
		},
		"comment": gin.H{
			"enabled":          h.cfg.Comment.Enabled,
			"require_approval": h.cfg.Comment.RequireApproval,
		},
		"social": gin.H{
			"github":  h.cfg.Social.GitHub,
			"twitter": h.cfg.Social.Twitter,
			"email":   h.cfg.Social.Email,
		},
		"custom_css": dbSettings["custom.css"],
		"custom_js":  dbSettings["custom.js"],
		"footer_text": dbSettings["footer.text"],
		"nav_items":   dbSettings["nav.items"],
	}

	c.JSON(http.StatusOK, gin.H{"settings": result})
}

// AdminUpdateSettings applies a batch of settings updates.
func (h *SettingsHandler) AdminUpdateSettings(c *gin.Context) {
	var updates map[string]string
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid request"}})
		return
	}

	if err := service.UpdateSettings(updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	// Apply theme changes to in-memory config for RSS feed etc.
	if v, ok := updates["site.title"]; ok {
		h.cfg.Site.Title = v
	}
	if v, ok := updates["site.subtitle"]; ok {
		h.cfg.Site.Subtitle = v
	}
	if v, ok := updates["site.description"]; ok {
		h.cfg.Site.Description = v
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

// PublicSettings returns site info and theme for the frontend.
func (h *SettingsHandler) PublicSettings(c *gin.Context) {
	dbSettings, _ := service.GetAllSettings()

	c.JSON(http.StatusOK, gin.H{
		"site": gin.H{
			"title":       or(dbSettings["site.title"], h.cfg.Site.Title),
			"subtitle":    or(dbSettings["site.subtitle"], h.cfg.Site.Subtitle),
			"description": or(dbSettings["site.description"], h.cfg.Site.Description),
		},
		"theme": gin.H{
			"name":         or(dbSettings["theme.name"], h.cfg.Theme.Name),
			"accent_color": or(dbSettings["theme.accent_color"], h.cfg.Theme.AccentColor),
			"font_body":    or(dbSettings["theme.font_body"], h.cfg.Theme.FontBody),
			"font_code":    or(dbSettings["theme.font_code"], h.cfg.Theme.FontCode),
		},
		"comment_enabled": h.cfg.Comment.Enabled,
		"social": gin.H{
			"github":  h.cfg.Social.GitHub,
			"twitter": h.cfg.Social.Twitter,
			"email":   h.cfg.Social.Email,
		},
		"custom_css": dbSettings["custom.css"],
		"footer_text": dbSettings["footer.text"],
	})
}

func or(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}
