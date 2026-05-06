package handler

import (
	"net/http"

	"mo/internal/database"
	"mo/internal/service"

	"github.com/gin-gonic/gin"
)

func GetPage(c *gin.Context) {
	slug := c.Param("slug")
	var content string
	err := database.DB.QueryRow("SELECT value FROM settings WHERE key = ?", "page:"+slug).Scan(&content)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Page not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"slug": slug, "content": content})
}

func UpdatePage(c *gin.Context) {
	slug := c.Param("slug")
	var body struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Content is required"}})
		return
	}

	// Render markdown to HTML
	htmlContent, err := service.RenderMarkdown(body.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to render markdown"}})
		return
	}

	if _, err := database.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", "page:"+slug, htmlContent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"slug": slug, "content": htmlContent})
}
