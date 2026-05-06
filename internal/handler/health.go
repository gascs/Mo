package handler

import (
	"net/http"

	"mo/internal/database"

	"github.com/gin-gonic/gin"
)

var Version = "dev"

func HealthCheck(c *gin.Context) {
	if err := database.DB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"version": Version,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": Version,
	})
}
