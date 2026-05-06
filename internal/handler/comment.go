package handler

import (
	"net/http"
	"strconv"

	"mo/internal/model"
	"mo/internal/service"

	"github.com/gin-gonic/gin"
)

// Public comments

type CreateCommentRequest struct {
	AuthorName  string `json:"author_name" binding:"required"`
	AuthorEmail string `json:"author_email" binding:"required"`
	AuthorURL   string `json:"author_url"`
	Content     string `json:"content" binding:"required"`
	ParentID    string `json:"parent_id"`
}

func GetComments(c *gin.Context) {
	slug := c.Param("slug")
	comments, err := service.GetCommentsByPostSlug(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

func CreateComment(c *gin.Context) {
	slug := c.Param("slug")

	// Rate limit
	ip := c.ClientIP()
	if !service.CheckCommentRateLimit(ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "RATE_LIMITED", "message": "Too many comments. Try again later."}})
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Name, email and content are required"}})
		return
	}

	comment, err := service.CreateComment(slug, req.ParentID, req.AuthorName, req.AuthorEmail, req.AuthorURL, req.Content, ip, c.GetHeader("User-Agent"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	if comment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Post not found"}})
		return
	}

	if comment.Status == "spam" {
		c.JSON(http.StatusOK, gin.H{"message": "Comment submitted for review", "status": "pending"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// Admin comments

func AdminListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	params := model.CommentListParams{
		Page:    page,
		PerPage: perPage,
		Status:  c.Query("status"),
		PostID:  c.Query("post_id"),
	}

	resp, err := service.ListComments(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func AdminUpdateCommentStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Status is required"}})
		return
	}

	validStatuses := map[string]bool{"approved": true, "spam": true, "trash": true, "pending": true}
	if !validStatuses[body.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid status"}})
		return
	}

	if err := service.UpdateCommentStatus(id, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment status updated"})
}

func AdminDeleteComment(c *gin.Context) {
	id := c.Param("id")
	if err := service.DeleteComment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}
