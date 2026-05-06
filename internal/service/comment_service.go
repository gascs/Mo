package service

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"mo/internal/database"
	"mo/internal/model"
)

var spamKeywords = []string{"viagra", "casino", "buy now", "click here", "free money", "http://", "www.", "赌博", "色情"}

var (
	commentRateStore = make(map[string]*commentRateEntry)
	commentRateMu    sync.Mutex
)

type commentRateEntry struct {
	count       int
	windowStart time.Time
}

func init() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			commentRateMu.Lock()
			now := time.Now()
			for k, v := range commentRateStore {
				if now.Sub(v.windowStart) > time.Minute {
					delete(commentRateStore, k)
				}
			}
			commentRateMu.Unlock()
		}
	}()
}

func CheckCommentRateLimit(ip string) bool {
	commentRateMu.Lock()
	defer commentRateMu.Unlock()
	now := time.Now()
	entry, exists := commentRateStore[ip]
	if !exists || now.Sub(entry.windowStart) > time.Minute {
		commentRateStore[ip] = &commentRateEntry{count: 1, windowStart: now}
		return true
	}
	entry.count++
	return entry.count <= 10
}

func CheckSpam(content string) bool {
	lower := strings.ToLower(content)
	for _, kw := range spamKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func CreateComment(postSlug, parentID, authorName, authorEmail, authorURL, content, ip, ua string) (*model.Comment, error) {
	// Find post ID from slug
	var postID string
	err := database.DB.QueryRow("SELECT id FROM posts WHERE slug = ? AND deleted_at IS NULL AND is_draft = 0", postSlug).Scan(&postID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Spam check
	isSpam := CheckSpam(content)
	status := "pending"
	if isSpam {
		status = "spam"
	}

	c := &model.Comment{
		ID:          model.NewULID(),
		PostID:      postID,
		ParentID:    parentID,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
		AuthorURL:   authorURL,
		Content:     content,
		Status:      status,
		UserAgent:   ua,
		IPAddress:   ip,
	}

	_, err = database.DB.Exec(
		"INSERT INTO comments (id, post_id, parent_id, author_name, author_email, author_url, content, status, user_agent, ip_address) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		c.ID, c.PostID, toNullString(c.ParentID), c.AuthorName, c.AuthorEmail, toNullString(c.AuthorURL), c.Content, c.Status, c.UserAgent, c.IPAddress,
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func toNullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func GetCommentsByPostSlug(slug string) ([]model.Comment, error) {
	rows, err := database.DB.Query(`
		SELECT c.id, c.post_id, c.parent_id, c.author_name, c.content, c.status, c.created_at
		FROM comments c JOIN posts p ON c.post_id = p.id
		WHERE p.slug = ? AND p.deleted_at IS NULL AND c.status = 'approved'
		ORDER BY c.created_at ASC`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		var parentID sql.NullString
		if err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.AuthorName, &c.Content, &c.Status, &c.CreatedAt); err != nil {
			continue
		}
		if parentID.Valid {
			c.ParentID = parentID.String
		}
		comments = append(comments, c)
	}
	if comments == nil {
		comments = []model.Comment{}
	}
	return comments, nil
}

func ListComments(params model.CommentListParams) (*model.CommentListResponse, error) {
	if params.Page <= 0 { params.Page = 1 }
	if params.PerPage <= 0 { params.PerPage = 20 }

	where := "WHERE 1=1"
	var args []interface{}

	if params.Status != "" {
		where += " AND c.status = ?"
		args = append(args, params.Status)
	}
	if params.PostID != "" {
		where += " AND c.post_id = ?"
		args = append(args, params.PostID)
	}

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM comments c "+where, args...).Scan(&total)

	offset := (params.Page - 1) * params.PerPage
	rows, err := database.DB.Query(`
		SELECT c.id, c.post_id, c.parent_id, c.author_name, c.author_email, c.author_url,
			c.content, c.status, c.ip_address, c.created_at
		FROM comments c `+where+` ORDER BY c.created_at DESC LIMIT ? OFFSET ?`,
		append(args, params.PerPage, offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		var parentID, authorURL sql.NullString
		if err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.AuthorName, &c.AuthorEmail,
			&authorURL, &c.Content, &c.Status, &c.IPAddress, &c.CreatedAt); err != nil {
			continue
		}
		if parentID.Valid { c.ParentID = parentID.String }
		if authorURL.Valid { c.AuthorURL = authorURL.String }
		comments = append(comments, c)
	}
	if comments == nil { comments = []model.Comment{} }

	return &model.CommentListResponse{
		Comments: comments,
		Total:    total,
		Page:     params.Page,
		PerPage:  params.PerPage,
	}, nil
}

func UpdateCommentStatus(id, status string) error {
	_, err := database.DB.Exec("UPDATE comments SET status = ? WHERE id = ?", status, id)
	return err
}

func DeleteComment(id string) error {
	_, err := database.DB.Exec("DELETE FROM comments WHERE id = ?", id)
	return err
}
