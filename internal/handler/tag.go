package handler

import (
	"net/http"

	"mo/internal/service"

	"github.com/gin-gonic/gin"
)

func ListTags(c *gin.Context) {
	tags, err := service.GetAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
