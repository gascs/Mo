package handler

import (
	"net/http"
	"strconv"

	"mo/internal/model"
	"mo/internal/service"

	"github.com/gin-gonic/gin"
)

type PostHandler struct{}

func NewPostHandler() *PostHandler {
	return &PostHandler{}
}

func (h *PostHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	params := model.PostListParams{
		Page:     page,
		PerPage:  perPage,
		Category: c.Query("category"),
		Status:   c.Query("status"),
		Tag:      c.Query("tag"),
		Search:   c.Query("search"),
	}

	resp, err := service.ListPosts(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PostHandler) Get(c *gin.Context) {
	id := c.Param("id")
	post, err := service.GetPost(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}
	if post == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Post not found"},
		})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) Create(c *gin.Context) {
	var req model.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid request"},
		})
		return
	}

	post, err := service.CreatePost(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid request"},
		})
		return
	}

	post, err := service.UpdatePost(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := service.SoftDeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post moved to trash"})
}

func (h *PostHandler) HardDelete(c *gin.Context) {
	id := c.Param("id")
	if err := service.HardDeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post permanently deleted"})
}

func (h *PostHandler) Restore(c *gin.Context) {
	id := c.Param("id")
	if err := service.RestorePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post restored"})
}

func (h *PostHandler) Publish(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Publish bool `json:"publish"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		body.Publish = true
	}

	if err := service.PublishPost(id, body.Publish); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}

	action := "published"
	if !body.Publish {
		action = "unpublished"
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post " + action})
}

func (h *PostHandler) Pin(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Pin bool `json:"pin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		body.Pin = true
	}

	if err := service.PinPost(id, body.Pin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post pin status updated"})
}

func (h *PostHandler) Dashboard(c *gin.Context) {
	stats, err := service.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, stats)
}
