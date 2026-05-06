package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"mo/internal/database"
	"mo/internal/model"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github-dark"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)
}

func RenderMarkdown(content string) (string, error) {
	var buf strings.Builder
	if err := md.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ExtractSummary(htmlContent string, maxChars int) string {
	// Strip HTML tags for plain text summary
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlContent, "")
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) > maxChars {
		return string(runes[:maxChars]) + "..."
	}
	return text
}

func GenerateSlug(title string) string {
	if title == "" {
		return fmt.Sprintf("post-%d", time.Now().UnixNano()/1000000)
	}

	slug := strings.ToLower(title)
	slug = strings.TrimSpace(slug)

	// Replace non-alphanumeric chars with dash
	re := regexp.MustCompile(`[^a-z0-9一-鿿]+`)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	// If only Chinese, generate timestamp-based slug
	if len(slug) == 0 || onlyChinese(slug) {
		slug = fmt.Sprintf("post-%d", time.Now().UnixNano()/1000000)
	}

	return slug
}

func onlyChinese(s string) bool {
	for _, r := range s {
		if r != '-' && !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return true
}

func CreatePost(req model.CreatePostRequest) (*model.Post, error) {
	if req.Category == "" {
		req.Category = "tech"
	}
	if req.Tags == "" {
		req.Tags = "[]"
	}

	slug := req.Slug
	if slug == "" {
		slug = GenerateSlug(req.Title)
	}

	// Ensure unique slug
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE slug = ? AND deleted_at IS NULL", slug).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		slug = slug + "-" + fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	}

	htmlContent, err := RenderMarkdown(req.Content)
	if err != nil {
		return nil, fmt.Errorf("render markdown: %w", err)
	}

	summary := req.Summary
	if summary == "" {
		summary = ExtractSummary(htmlContent, 200)
	}

	isDraft := 1
	if !req.IsDraft {
		isDraft = 0
	}

	post := &model.Post{
		ID:          model.NewULID(),
		Title:       req.Title,
		Slug:        slug,
		Content:     req.Content,
		ContentHTML: htmlContent,
		Summary:     summary,
		Category:    req.Category,
		Tags:        req.Tags,
		IsDraft:     isDraft == 1,
	}

	var publishedAt interface{}
	if !post.IsDraft {
		now := time.Now()
		post.PublishedAt = &now
		publishedAt = now
	} else {
		publishedAt = nil
	}

	_, err = database.DB.Exec(`
		INSERT INTO posts (id, title, slug, content, content_html, summary, category, tags, is_pinned, is_draft, published_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		post.ID, post.Title, post.Slug, post.Content, post.ContentHTML, post.Summary,
		post.Category, post.Tags, isDraft, publishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert post: %w", err)
	}

	return post, nil
}

func UpdatePost(id string, req model.UpdatePostRequest) (*model.Post, error) {
	// Fetch existing post
	existing, err := GetPost(id)
	if err != nil {
		return nil, err
	}

	// Apply changes
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Content != "" {
		existing.Content = req.Content
		htmlContent, err := RenderMarkdown(req.Content)
		if err != nil {
			return nil, fmt.Errorf("render markdown: %w", err)
		}
		existing.ContentHTML = htmlContent
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Tags != "" {
		existing.Tags = req.Tags
	}
	if req.IsDraft != nil {
		existing.IsDraft = *req.IsDraft
		if !existing.IsDraft && existing.PublishedAt == nil {
			now := time.Now()
			existing.PublishedAt = &now
		}
	}
	if req.IsPinned != nil {
		existing.IsPinned = *req.IsPinned
	}
	if req.IsPrivate != nil {
		existing.IsPrivate = *req.IsPrivate
	}
	if req.Summary != "" {
		existing.Summary = req.Summary
	} else if req.Content != "" {
		existing.Summary = ExtractSummary(existing.ContentHTML, 200)
	}

	if req.Slug != "" && req.Slug != existing.Slug {
		var count int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE slug = ? AND id != ? AND deleted_at IS NULL", req.Slug, id).Scan(&count)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			req.Slug = req.Slug + "-" + fmt.Sprintf("%d", time.Now().UnixNano()%10000)
		}
		existing.Slug = req.Slug
	}

	isDraft := 0
	if existing.IsDraft {
		isDraft = 1
	}
	isPinned := 0
	if existing.IsPinned {
		isPinned = 1
	}
	isPrivate := 0
	if existing.IsPrivate {
		isPrivate = 1
	}

	_, err = database.DB.Exec(`
		UPDATE posts SET title=?, slug=?, content=?, content_html=?, summary=?, category=?, tags=?,
		is_pinned=?, is_draft=?, is_private=?, published_at=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=? AND deleted_at IS NULL`,
		existing.Title, existing.Slug, existing.Content, existing.ContentHTML, existing.Summary,
		existing.Category, existing.Tags, isPinned, isDraft, isPrivate, existing.PublishedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}

	return existing, nil
}

func GetPost(id string) (*model.Post, error) {
	var post model.Post
	var publishedAt, deletedAt sql.NullTime
	var tags string

	err := database.DB.QueryRow(`
		SELECT id, title, slug, content, content_html, summary, category, tags,
			is_pinned, is_draft, is_private, published_at, created_at, updated_at, deleted_at
		FROM posts WHERE id = ?`, id,
	).Scan(&post.ID, &post.Title, &post.Slug, &post.Content, &post.ContentHTML,
		&post.Summary, &post.Category, &tags, &post.IsPinned, &post.IsDraft,
		&post.IsPrivate, &publishedAt, &post.CreatedAt, &post.UpdatedAt, &deletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	post.Tags = tags
	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	}
	if deletedAt.Valid {
		post.DeletedAt = &deletedAt.Time
	}
	return &post, nil
}

func ListPosts(params model.PostListParams) (*model.PostListResponse, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 {
		params.PerPage = 10
	}

	where := "WHERE deleted_at IS NULL"
	args := []interface{}{}

	if params.Category != "" {
		where += " AND category = ?"
		args = append(args, params.Category)
	}

	switch params.Status {
	case "published":
		where += " AND is_draft = 0"
	case "draft":
		where += " AND is_draft = 1"
	case "trashed":
		where = "WHERE deleted_at IS NOT NULL"
	}

	if params.Tag != "" {
		where += " AND tags LIKE ?"
		args = append(args, `%"`+params.Tag+`"%`)
	}

	if params.Search != "" {
		where += " AND (title LIKE ? OR content LIKE ?)"
		s := "%" + params.Search + "%"
		args = append(args, s, s)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM posts " + where
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.PerPage
	query := `SELECT id, title, slug, content, content_html, summary, category, tags,
		is_pinned, is_draft, is_private, published_at, created_at, updated_at, deleted_at
		FROM posts ` + where + ` ORDER BY is_pinned DESC, created_at DESC LIMIT ? OFFSET ?`

	queryArgs := append(args, params.PerPage, offset)
	rows, err := database.DB.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var p model.Post
		var tags string
		var publishedAt, deletedAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.ContentHTML,
			&p.Summary, &p.Category, &tags, &p.IsPinned, &p.IsDraft,
			&p.IsPrivate, &publishedAt, &p.CreatedAt, &p.UpdatedAt, &deletedAt,
		); err != nil {
			return nil, err
		}
		p.Tags = tags
		if publishedAt.Valid {
			p.PublishedAt = &publishedAt.Time
		}
		if deletedAt.Valid {
			p.DeletedAt = &deletedAt.Time
		}
		posts = append(posts, p)
	}

	if posts == nil {
		posts = []model.Post{}
	}

	return &model.PostListResponse{
		Posts:   posts,
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
	}, nil
}

func SoftDeletePost(id string) error {
	_, err := database.DB.Exec("UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

func HardDeletePost(id string) error {
	_, err := database.DB.Exec("DELETE FROM posts WHERE id = ?", id)
	return err
}

func RestorePost(id string) error {
	_, err := database.DB.Exec("UPDATE posts SET deleted_at = NULL WHERE id = ?", id)
	return err
}

func PublishPost(id string, publish bool) error {
	if publish {
		_, err := database.DB.Exec("UPDATE posts SET is_draft = 0, published_at = COALESCE(published_at, CURRENT_TIMESTAMP), updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
		return err
	}
	_, err := database.DB.Exec("UPDATE posts SET is_draft = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

func PinPost(id string, pin bool) error {
	v := 0
	if pin {
		v = 1
	}
	_, err := database.DB.Exec("UPDATE posts SET is_pinned = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", v, id)
	return err
}

func GetAllTags() ([]model.TagItem, error) {
	rows, err := database.DB.Query("SELECT tags FROM posts WHERE deleted_at IS NULL AND is_draft = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagCount := make(map[string]int)
	for rows.Next() {
		var tagsJSON string
		if err := rows.Scan(&tagsJSON); err != nil {
			continue
		}
		var tags []string
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			continue
		}
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if t != "" {
				tagCount[t]++
			}
		}
	}

	var result []model.TagItem
	for name, count := range tagCount {
		result = append(result, model.TagItem{Name: name, Count: count})
	}
	if result == nil {
		result = []model.TagItem{}
	}
	return result, nil
}

func GetDashboardStats() (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}
	row := database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL")
	row.Scan(&stats.TotalPosts)
	row = database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL AND is_draft = 0")
	row.Scan(&stats.PublishedPosts)
	row = database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL AND is_draft = 1")
	row.Scan(&stats.DraftPosts)
	row = database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL AND category = 'treehole'")
	row.Scan(&stats.TreeholePosts)
	return stats, nil
}
