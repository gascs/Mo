package handler

import (
	"net/http"
	"strconv"

	"mo/internal/service"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	UploadDir string
}

func NewMediaHandler(uploadDir string) *MediaHandler {
	return &MediaHandler{UploadDir: uploadDir}
}

func (h *MediaHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "No file provided"},
		})
		return
	}

	media, err := service.UploadFile(file, h.UploadDir)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "UPLOAD_ERROR", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, media)
}

func (h *MediaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	resp, err := service.ListMedia(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *MediaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteMedia(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Media deleted"})
}
