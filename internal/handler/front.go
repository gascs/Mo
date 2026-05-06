package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"mo/internal/database"
	"mo/internal/model"

	"github.com/gin-gonic/gin"
)

// PublicPostList returns published, non-deleted posts for the frontend.
func PublicPostList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	category := c.Query("category")
	tag := c.Query("tag")

	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 50 {
		perPage = 10
	}

	where := "WHERE deleted_at IS NULL AND is_draft = 0"
	var args []interface{}

	if category != "" {
		where += " AND category = ?"
		args = append(args, category)
	}
	if tag != "" {
		where += " AND tags LIKE ?"
		args = append(args, `%"`+tag+`"%`)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM posts " + where
	if err := database.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	offset := (page - 1) * perPage
	query := `SELECT id, title, slug, summary, category, tags, is_pinned, published_at, created_at
		FROM posts ` + where + ` ORDER BY is_pinned DESC, published_at DESC LIMIT ? OFFSET ?`

	queryArgs := append(args, perPage, offset)
	rows, err := database.DB.Query(query, queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	defer rows.Close()

	type PostItem struct {
		ID          string     `json:"id"`
		Title       string     `json:"title"`
		Slug        string     `json:"slug"`
		Summary     string     `json:"summary"`
		Category    string     `json:"category"`
		Tags        string     `json:"tags"`
		IsPinned    bool       `json:"is_pinned"`
		PublishedAt *string    `json:"published_at"`
		CreatedAt   string     `json:"created_at"`
	}

	var posts []PostItem
	for rows.Next() {
		var p PostItem
		var tags string
		var publishedAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.Category, &tags, &p.IsPinned, &publishedAt, &p.CreatedAt); err != nil {
			continue
		}
		p.Tags = tags
		if publishedAt.Valid {
			s := publishedAt.Time.Format("2006-01-02")
			p.PublishedAt = &s
		}
		posts = append(posts, p)
	}
	if posts == nil {
		posts = []PostItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":    posts,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// PublicPostBySlug returns a single published post by slug.
func PublicPostBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var p model.Post
	var tags string
	var publishedAt, deletedAt sql.NullTime

	err := database.DB.QueryRow(`
		SELECT id, title, slug, content, content_html, summary, category, tags,
			is_pinned, is_draft, published_at, created_at, updated_at
		FROM posts WHERE slug = ? AND deleted_at IS NULL AND is_draft = 0`, slug,
	).Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.ContentHTML,
		&p.Summary, &p.Category, &tags, &p.IsPinned, &p.IsDraft,
		&publishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Post not found"}})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	p.Tags = tags
	if publishedAt.Valid {
		p.PublishedAt = &publishedAt.Time
	}
	if deletedAt.Valid {
		p.DeletedAt = &deletedAt.Time
	}

	c.JSON(http.StatusOK, p)
}

// SearchPosts performs full-text search via FTS5.
func SearchPosts(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, gin.H{"posts": []interface{}{}, "total": 0})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if page <= 0 { page = 1 }
	if perPage <= 0 { perPage = 10 }

	likeQ := "%" + q + "%"
	rows, err := database.DB.Query(`
		SELECT id, title, slug, summary, category, published_at, created_at
		FROM posts WHERE deleted_at IS NULL AND is_draft = 0 AND (title LIKE ? OR content LIKE ?)
		ORDER BY published_at DESC LIMIT ? OFFSET ?`,
		likeQ, likeQ, perPage, (page-1)*perPage,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	defer rows.Close()

	type SearchItem struct {
		ID          string  `json:"id"`
		Title       string  `json:"title"`
		Slug        string  `json:"slug"`
		Summary     string  `json:"summary"`
		Category    string  `json:"category"`
		PublishedAt *string `json:"published_at"`
		CreatedAt   string  `json:"created_at"`
	}

	var posts []SearchItem
	for rows.Next() {
		var s SearchItem
		var pub sql.NullTime
		if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Summary, &s.Category, &pub, &s.CreatedAt); err != nil {
			continue
		}
		if pub.Valid {
			str := pub.Time.Format("2006-01-02")
			s.PublishedAt = &str
		}
		posts = append(posts, s)
	}
	if posts == nil {
		posts = []SearchItem{}
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts, "total": len(posts), "page": page, "per_page": perPage})
}

// Archive returns posts grouped by year and month.
func Archive(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, title, slug, category, published_at, created_at
		FROM posts WHERE deleted_at IS NULL AND is_draft = 0
		ORDER BY published_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	defer rows.Close()

	type ArchiveItem struct {
		ID          string  `json:"id"`
		Title       string  `json:"title"`
		Slug        string  `json:"slug"`
		Category    string  `json:"category"`
		PublishedAt *string `json:"published_at"`
	}

	type YearGroup struct {
		Year  string        `json:"year"`
		Posts []ArchiveItem `json:"posts"`
	}

	groups := make(map[string][]ArchiveItem)
	var years []string

	for rows.Next() {
		var item ArchiveItem
		var pub sql.NullTime
		var createdAt string
		if err := rows.Scan(&item.ID, &item.Title, &item.Slug, &item.Category, &pub, &createdAt); err != nil {
			continue
		}
		dateStr := createdAt
		if pub.Valid {
			dateStr = pub.Time.Format("2006-01-02")
			item.PublishedAt = &dateStr
		}
		year := dateStr[:4]
		if _, exists := groups[year]; !exists {
			years = append(years, year)
		}
		groups[year] = append(groups[year], item)
	}

	var result []YearGroup
	for _, year := range years {
		result = append(result, YearGroup{Year: year, Posts: groups[year]})
	}
	if result == nil {
		result = []YearGroup{}
	}

	c.JSON(http.StatusOK, gin.H{"archive": result})
}
